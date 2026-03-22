---
title: Container Orchestration Golden Path
date: 2024-10-31
description: Best practises for building and orchestrating containers with Kubernetes.
tags:
  - tech
  - how-to
---

This document outlines best practices for building and orchestrating containers using Kubernetes.

## Definitions

| Term               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| :----------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Container**      | Containers are packages of software that contain all of the necessary elements to run in any environment. They virtualise the operating system and enable the application to run in the same manner anywhere, from a private data center or public cloud, or even on the developer's personal laptop.                                                                                                                                              |
| **Docker**         | A platform as a service (PaaS) that provides a set of software tools and services for building, shipping, and running containerized applications. It allows developers to package their applications and dependencies into portable containers that can be run on any system with an OCI-compliant runtime, such as Kubernetes or Docker Engine.                                                                                                   |
| **Kubernetes**     | Kubernetes is an open-source container orchestration system for automating software deployment, scaling, and management. It assembles one or more computers, either virtual machines or bare metal, into a cluster which can run workloads in containers. The brains of the system consists of a control-plane that runs on the “master” node/s, and features a reconciliation loop that drives the actual cluster state toward the desired state. |
| **SIGTERM**        | A generic signal used to cause program termination.                                                                                                                                                                                                                                                                                                                                                                                                |
| **Kubelet**        | The agent that runs on each node in a Kubernetes cluster. It communicates with the Kubernetes control plane and is responsible for running containers, managing pods, and ensuring that the desired state of the cluster is maintained at the node-level.                                                                                                                                                                                          |
| **Deployment**     | A Deployment manages a set of Pods to run an application workload, usually one that doesn't maintain state.                                                                                                                                                                                                                                                                                                                                        |
| **ServiceAccount** | A ServiceAccount is a type of non-human account that, in Kubernetes, provides a distinct identity in a Kubernetes cluster. Application Pods, system components, and entities inside and outside the cluster can use a specific ServiceAccount's credentials to identify as that ServiceAccount.                                                                                                                                                    |
| **ConfigMap**      | A ConfigMap is a Kubernetes resource that stores non-sensitive configuration data, such as application settings and environment variables. ConfigMaps can be used to provide configuration information to containers.                                                                                                                                                                                                                              |
| **Secret**         | Similar to a ConfigMap, a Secret is a Kubernetes resource that stores sensitive information, such as passwords, API keys, and certificates.                                                                                                                                                                                                                                                                                                        |
| **ExternalSecret** | ExternalSecret is a resource that allows you to manage secrets stored in external secret management systems, such as HashiCorp Vault, AWS Secrets Manager or AWS SSM Parameter Store, within your Kubernetes cluster. ExternalSecret provides a way to securely store and manage secrets outside of Kubernetes while still making them accessible to your applications.                                                                            |
| **HPA**            | HorizontalPodAutoscaler (HPA for short) automatically updates a workload resource (such as a Deployment or StatefulSet), with the aim of automatically scaling the workload to match demand.                                                                                                                                                                                                                                                       |
| **IAM**            | Identity and Access Management (IAM) manages Amazon Web Services (AWS) users and their access to AWS accounts and services. It controls the level of access a user can have over an AWS account & set users, grant permission, and allows a user to use different features of an AWS account.                                                                                                                                                      |
| **Trust policy**   | A JSON policy document in which you define the principals that you trust to assume the role. A role trust policy is a required resource-based policy that is attached to a role in IAM. The principals that you can specify in the trust policy include users, roles, accounts, and services.                                                                                                                                                      |

## What is Kubernetes? And why?

Nothing explains Kubernetes and its benefits better than the [Google Kubernetes Engine (GKE) comic by Scott McCloud](https://cloud.google.com/kubernetes-engine/kubernetes-comic).

It is required reading, mandatory for all, regardless of your level of expertise with containers or Kubernetes.

If you haven’t read it, go do so and then return. If you have, read it again and then return.

Despite a slight learning curve, Kubernetes is by far the easiest and most effective way to orchestrate your applications in a way that makes them portable, reproducible and scalable while reducing the possibility of downtime.

## Core Guidelines

### Principle of Least Bloat (PoLB)

Priority number one is to remove any bloat in your build. Containers work best when they are lightweight and portable, with only the components required to run the app–nothing more.

There are 4 main reasons for this:

1. **More secure** (fewer components means fewer vulnerabilities).
2. **Portability** (faster builds, transfers, and faster startups leading to faster deployment & scaling).
3. **More efficient** (requires fewer system resources, wherever it is ran).
4. **Lower cost** (less storage and system resources are consumed, resulting in lower cost).

The ultimate outcome is a container built with the most minimal base image possible, with a single binary to run your application.

This is often not possible and the application needs more, but the goal is to ensure that anything that ends up on your container is actually required to run your application in production. Anything else is just bloat and impacts the security, portability, efficiency and cost of your container.

To minimise bloat, you can:

- Choose a minimal base image, e.g. [scratch](https://hub.docker.com/_/scratch) or a [Chainguard](https://www.chainguard.dev/chainguard-images) image.
- Ignore unused files (use a strict `.dockerignore` file and selectively copy only what is needed).
- Use multi-stage builds, only copying the build files to the final stage.
- Manage cache files (e.g. `apk`, `apt` or `npm` caches should not make it into the final build).

### Logging

Docker and Kubernetes automatically capture `stdout` and `stderr` streams from your containers. In the case of Kubernetes, the logs are captured for all containers and are made available via `kubectl`.

Additionally, tools are available that allow you to ship them to a centralised location where you can analyse them effectively. This is the preferred mechanism for logs management.

Thankfully, this means that all you need to do is log to `stdout` and `stderr`.

Ensure you are not logging to a file in addition to `stdout` and `stderr`, as this creates a plethora of issues and unnecessary complexity, including:

- Log file management and rotation becomes crucial and a single point of failure.
- In environments with limited disk space or I/O, writing to a file can introduce performance overhead.
- Scaling becomes harder, or impossible (e.g. many containers write to a discrete log file, or they mount a shared volume and attempt to write to the same log file).
- Security concerns. If the container is compromised, the file may contain sensitive information, therefor it increases the attack vector of the application.

Containers are supposed to be ephemeral. Any state that you wish to store is ideally shipped elsewhere, and this includes logs.

### Healthchecks

Healthchecks are essential for Kubernetes to monitor the health of pods and ensure that only healthy instances are serving traffic. There are two primary types of healthchecks: readiness probes and liveness probes.

**Readiness Probes**

- **Purpose:** Determine if a pod is ready to receive traffic.
- **Used:** For applications that require initialisation or warm-up time before they are fully functional, e.g. any app that needs to connect to a dependency.

**Liveness Probes**

- **Purpose:** Determine if a pod is still alive and functioning as expected.
- **Used:** To detect issues such as application crashes, deadlocks, or resource exhaustion.

Both endpoints should be built into any application that is intended for a container.

**Example endpoints:**

```javascript
const express = require("express");
const app = express();

app.get("/readyz", (req, res) => {
  // Check application readiness, including dependencies
  const isReady = checkDatabaseConnection();
  if (isReady) {
    res.status(200).json({ status: "READY" });
  } else {
    res.status(503).json({ status: "SERVICE_UNAVAILABLE" });
  }
});

app.get("/healthz", (req, res) => {
  // Simply respond to verify that the server able
  res.status(200).json({ status: "OK" });
});
```

Often the `healthz` endpoint simply responds with a 200 (e.g. no `checkThing()` is performed), which simply verifies that the server is running and able to respond (no crashes, deadlocks or resource exhaustion). However, the need for checks will change with each case and it is important to consider the system as a whole.

On one hand, your use-case may demand that you kill the application if a certain service becomes unavailable. On the other hand, being too liberal and checking each and every external dependency could lead to catastrophic failures from cascading unavailability.

Cascading unavailability? Sounds like some fancy term created by "DevOps" to make someone sound smart, but let me explain with an example…

- **Backend Failure:** The backend service experiences a temporary outage due to a network issue or application error.
- **Frontend Health Check Failure:** The frontend service's health check fails to reach the backend.
- **Pod Termination:** Kubernetes detects the health check failure and terminates the frontend pod.
- **Pod Replacement:** Kubernetes attempts to replace the terminated pod with a new one, but the underlying issue with the backend may prevent this from happening immediately.
- **User Experience Impact:** Users are unable to access the frontend service, and as a result we are unable to give them a descriptive error message or feedback.

Unavailability has cascaded... 1 level (when the frontend checked the backend). This is a simple (and stupid) example, but keep adding dependency checks willy nilly and you will end up with a perfectly beautiful waterfall of unavailable services.

**Example usage:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 0
            periodSeconds: 10
            timeoutSeconds: 1
          livenessProbe:
            httpGet:
              path: /readyz
              port: 8080
            initialDelaySeconds: 120
            periodSeconds: 30
            timeoutSeconds: 1
```

### Graceful Shutdowns

All applications should listen for `SIGTERM` signals and handle shut-down appropriately.

A `SIGTERM` is a generic signal used to cause program termination.

In the case of Kubernetes, when a container needs to be terminated, a `SIGTERM` signal is sent by the `kubelet` to cause the main process of the container to terminate.

Cases where this occurs are:

- A new deployment rollout
- A scale-down event

It is important that the `SIGTERM` is handled by your application to ensure any outstanding operations are completed without error. This is never implemented for you by any libraries, or at least if it is, it won’t be tailored to your use-case.

Once a `SIGTERM` is received, the application should begin to shut down the gracefully, finalising any remaining requests to avoid an error response or an unintended response from erroneous completion of operations.

The basic ideas is to:

- Allow running transactions to complete
- Do any application cleanup required
- Stop accepting new connections (while Kubernetes handles this as well, it’s good practise to do so in the app as well)
- Close any inactive connections (keep alive requests, websockets)

**Do**

```javascript
// Run cleanup() and exit gracefully on sigterm
const gracefulShutdown = () => {
  console.log("Received kill signal, shutting down gracefully");
  server.close(() => {
    database.close(); // Shut down the db
    cleanup(); // Perform any other necessary cleanup tasks
    console.log("Closed out remaining connections - exiting gracefully");
    process.exit(0);
  });

  setTimeout(() => {
    console.error(
      "Could not close connections in time, forcefully shutting down",
    );
    process.exit(1);
  }, 10000); // Give it a 10s deadline
};

process.on("SIGTERM", gracefulShutdown);
process.on("SIGINT", gracefulShutdown); // Also shutdown on sigint (e.g. ctrl-c)
```

**Don’t**

```javascript
// Use `sleep` to delay shutdown on sigterm
process.on("SIGTERM", () => {
  console.log("Received kill signal, shutting down not so gracefully");
  setTimeout(() => {
    console.log("Exiting");
    process.exit(0);
  }, 5000); // Sleep for 5 seconds
});
```

```yaml
// Use preStop to run 'sleep 10' before shutting down
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          ports:
            - containerPort: 8080
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 10"] // Sleep 10s before shutdown
```

### Environment Variables and Secrets

#### Use a Secret for Sensitive Environment Variables

Avoid exposing sensitive values through mis-use of ConfigMap or raw env values in your Kubernetes manifests.

Anything that is sensitive should go in a Kubernetes Secret resource. Otherwise, it is fine to put it in a ConfigMap resource.

**Do**

```yaml
// Store sensitive values in a Secret
apiVersion: v1
kind: Secret
metadata:
  name: my-config
data:
  my-sensitive-var: "sensitive_value"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          env:
            - name: MY_SENSITIVE_VAR
              valueFrom:
                secretKeyRef:
                  name: my-config
                  key: my-sensitive-var
```

**Don’t**

```yaml
// Store sensitive values in the deployment manifest
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          env:
            - name: MY_SENSITIVE_VAR
              value: "sensitive_value"
```

```yaml
// Store sensitive values in a ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  my-sensitive-var: "sensitive_value"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          env:
            - name: MY_SENSITIVE_VAR
              valueFrom:
                configMapRef:
                  name: my-config
                  key: my-sensitive-var
```

You may argue that a Secret resource is simply base64 encoded and not secure, but the important difference is the access controls that are applied to each resource type.

#### Use ExternalSecret’s for Secrets Generation

An [ExternalSecret](https://external-secrets.io/v0.4.4/api-externalsecret/) is a resource that describes what data should be fetched, how the data should be transformed and saved as a standard Kubernetes Secret resources.

It allows you to store the value of the secret in your preferred secret manager (e.g. AWS SSM Parameter Store), and store the Kubernetes configuration for the secret in a non-sensitive way.

#### Supply Helm Charts with an Existing Secret

Most public Helm charts allow you to specify an existing secret that contains sensitive environment variables in the format that it expects. This is the preferred pattern when using public charts, as well as when developing internal Helm charts.

**Example Helm template:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: { { include "my-app.fullname" . } }
spec:
  template:
    metadata:
      labels:
        app: { { include "my-app.fullname" . } }
    spec:
      containers:
        - name: { { include "my-app.fullname" . } }
          image: { { include "my-app.image" . } }
          imagePullPolicy: { { .Values.image.pullPolicy | quote } }
          env:
            - name: PASSWORD
              valueFrom:
                secretKeyRef:
                  name: { { include "my-app.secretName" . } }
                  key: password
```

**Example values.yaml:**

```yaml
secretName: my-app-secrets
```

**Example ExternalSecret:**

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-app-secrets
  namespace: my-app
spec:
  secretStoreRef:
    kind: ClusterSecretStore
    name: cluster-parameter-store
  refreshInterval: 1h
  target:
    name: my-app-secrets
    creationPolicy: Owner
  data:
    - secretKey: password
      remoteRef:
        key: /my-app/password
```

### Namespaces

Namespaces allow for the logical isolation of resources, helping us configure targeted permissions and access controls, as well as to provide structure for other configurations (e.g. observability)

Rule #1 in Kubernetes security best practices is to avoid using the `default` namespace.

Here are some of the primary concerns:

- **Lack of isolation:** All resources created in the default namespace share the same permissions and access controls. This can lead to unintended access or conflicts between different applications or teams.
- **Increased attack surface:** The default namespace is often used by system components and critical applications, making it a prime target for attackers. A breach in the default namespace could have widespread consequences.
- **Difficulty in managing permissions:** Managing permissions and access controls in the default namespace can be complex and error-prone, as it involves granting or revoking permissions for a large number of resources.
- **Limited visibility:** It can be challenging to monitor and track resource usage in the default namespace, making it difficult to identify and address potential security issues.

A namespace per app is the generally accepted standard.

**Do**

```yaml
// Group namespaces by application
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-backend
  namespace: enterprise-backend
```

**Don’t**

```yaml
// Put everything in the default namespace
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-backend
  namespace: default
```

```yaml
// Group namespaces by an overarching product
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-backend
  namespace: enterprise
```

### Volume Mounts

Volume mounts allow containers to access files or directories from the host filesystem or other storage sources.

**Best Practices:**

- **Use Persistent Volumes (PVs) for persistent data:** PVs ensure that data persists even if pods are restarted or rescheduled. Use Persistent Volume Claims (PVCs) to request PVs within your deployments or stateful sets.
- **Avoid hostPath volume type:** Using hostPath exposes the hosts filesystem to the container, which presents many security risks. If you can avoid using a hostPath volume, you should. For example, define a local PersistentVolume, and use that instead. However, avoid using either if you can.
- **Use read-only mounts:** If your application only needs to read data, use read-only mounts to improve performance and security.
- **Optimize volume size:** Avoid creating excessively large volumes, as this can waste storage resources.

### Resource Limits and Requests

For smooth operation of our container, we need to set its resource Requests and Limits. This ensures that a sufficient allocation of resources are set aside for it to run.

In most cases, setting the resource Requests and Limits is also required before we can auto-scale its Deployment replicas (more on this next).

Requests and Limits are the two key concepts in Kubernetes that define the resource allocation for a container.

**Requests:**

- Specify the **minimum** amount of resources (CPU and memory) that a container needs to function properly.
- If a container requests more resources than are available, Kubernetes will attempt to schedule it on another node with sufficient resources.
- Requests are used for scheduling and resource allocation decisions.

**Limits:**

- Specify the **maximum** amount of resources a container can use.
- If a container exceeds its resource limits, Kubernetes may take actions to throttle the container.
- Limits are used to prevent containers from consuming excessive resources and impacting other applications in the cluster.

**Relationship Between Requests and Limits:**

- **Requests must be less than or equal to limits.**
- If a container's request is less than its limit, Kubernetes can scale it up if necessary to meet the requested resources (AKA burstable).
- If a container's request is equal to its limit, Kubernetes will try to allocate the exact amount of resources requested.

While the correct allocation strategy will vary per workload, a good starting point is to work out a somewhat liberal memory allocation for the requests, then set limit == request.

This will ensure each app gets its share of available memory, avoiding OOM kills. We can allow the CPU to burst (e.g. leave the limit blank). CPU is a fundamentally different resource to memory and can handle being passed around from pod to pod, and pods are happy to wait for it when times are tough.

To understand more on this, see the following article:
[https://home.robusta.dev/blog/kubernetes-memory-limit](https://home.robusta.dev/blog/kubernetes-memory-limit)

**Example:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          resources:
            requests:
              memory: "4Gi"
              cpu: "250m"
            limits:
              memory: "4Gi" # memory limit == request
              # no cpu limit - let it spike!
```

## Additional Guidelines

### Horizontal Scaling

Horizontal Scaling refers to increasing or decreasing the number of instances of a running application. This is achieved by adding or removing pods in a Kubernetes cluster.

Horizontal scaling is limited in on-premises environments due to the constraints of physical infrastructure. Adding additional nodes to a cluster is generally not possible. Therefore, optimising your applications for concurrency is essential to ensure efficient performance and handle increasing workloads without relying too heavily on horizontal scaling.

#### HPA (HorizontalPodAutoscaler)

Now we have our resource requests set (as above), we can use these to inform HPA (HorizontalPodAutoscaler) of replica scaling decisions for our deployment.

Using HPA (HorizontalPodAutoscaler) to scale deployments based on a target average utilisation of CPU request is the most basic form of scaling there is, but it is simple and perfectly sufficient for 99% of situations.

Similar targets can be configured for memory, or both cpu and memory at once.

**Example**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

The above manifest tells HPA (HorizontalPodAutoscaler) that a minimum of 2 replicas should be configured for redundancy at any one time, and a maximum of 10 can be scheduled at any one time.

If the average CPU request utilisation for all replicas in the deployment reaches the target average utilisation of 70%, the replicas for the deployment should be scaled up to match demand (and vice versa).

It is important not to explicitly set your replica count via the Deployment manifest if you are managing your replicas via HPA, as a deployment rollout might override what HPA has previously set and cause a scale-down event for your workload.

**Do**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  # "replicas" is omitted and controlled by HPA
  template:
    ...
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  ...
```

**Don’t**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 2 # <-- !! will override what HPA sets
  template:
    ...
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  ...
```

#### KEDA (Kubernetes-based Event Driven Autoscaler)

If HPA doesn’t cover your use-case, KEDA steps in, which allows you to scale based on any internal or external metric you like using [existing Scalers](https://keda.sh/docs/2.14/scalers/), or a custom Scaler.

When choosing a trigger for your scaling decisions, it is important to consider the time it takes from a scale-up event to a running workload. Scaling events will always take at least 30s in any environment and longer to produce a running replica, so relying on a metric/threshold that is too fine-grained does not scale well (pun accidental).

**Do**

```yaml
// Scale my-app using http_requests_total of 100/2m from prometheus
apiVersion: keda.k8s.io/v1beta1
kind: ScaledObject
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  pollingInterval: 30s
  cooldownPeriod: 30s
  minReplicaCount: 1
  maxReplicaCount: 10
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus-server:9090
        query: sum(rate(http_requests_total{deployment="my-app"}[2m]))
        threshold: '100'
        activationThreshold: '5.5'
```

**Don’t**

```yaml
// Scale my-app using http_requests_total of 1/1m from prometheus
apiVersion: keda.k8s.io/v1beta1
kind: ScaledObject
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  pollingInterval: 30s
  cooldownPeriod: 30s
  minReplicaCount: 1
  maxReplicaCount: 10
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus-server:9090
        query: sum(rate(http_requests_total{deployment="my-app"}[1m]))
        threshold: '1'
        activationThreshold: '1'
```

### Service Accounts and AWS IAM Roles

Amazon EKS supports using OpenID Connect (OIDC) identity providers as a method to authenticate users to your cluster.

This removes the need for storing your AWS secret access keys in container environment variables, and allows you to simply scope access to the role to a specific Kubernetes ServiceAccount resource in a specific Namespace.

**Do**

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  annotations:
    # The pre-configured AWS IAM role arn
    eks.amazonaws.com/role-arn: "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  serviceAccountName: my-service-account
  # ... other pod specifications
```

**Don’t**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: my-image:latest
          env:
            - name: AWS_ACCESS_KEY_ID
              value: "your_access_key_id"
            - name: AWS_SECRET_ACCESS_KEY
              value: "your_secret_access_key"
```

To allow your IAM role to be assumed, a trust policy is required, which defines the principals that you _trust_ to assume the role (e.g. the `my-service-account` ServiceAccount in the `my-namespace` Namespace).

**Example**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/YOUR_OIDC_PROVIDER_URL"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "YOUR_OIDC_PROVIDER_URL/id/my-service-account": "system:serviceaccount:my-namespace:my-service-account",
          "YOUR_OIDC_PROVIDER_URL/id/my-service-account": "sts.amazonaws.com"
        }
      }
    }
  ]
}
```

## Additional Resources

- [https://docs.docker.com/build/building/best-practices/](https://docs.docker.com/build/building/best-practices/)
- [https://kubernetes.io/docs/concepts/cluster-administration/logging/](https://kubernetes.io/docs/concepts/cluster-administration/logging/)
- [https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [https://kubernetes.io/docs/concepts/configuration/configmap/](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [https://kubernetes.io/docs/concepts/configuration/secret/](https://kubernetes.io/docs/concepts/configuration/secret/)
- [https://external-secrets.io/v0.4.4/api-externalsecret/](https://external-secrets.io/v0.4.4/api-externalsecret/)
- [https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [https://kubernetes.io/docs/concepts/storage/volumes/](https://kubernetes.io/docs/concepts/storage/volumes/)
- [https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [https://keda.sh/docs/latest](https://keda.sh/docs/latest)
- [https://kubernetes.io/docs/concepts/security/service-accounts/](https://kubernetes.io/docs/concepts/security/service-accounts/)
- [https://aws.amazon.com/blogs/containers/introducing-oidc-identity-provider-authentication-amazon-eks/](https://aws.amazon.com/blogs/containers/introducing-oidc-identity-provider-authentication-amazon-eks/)
