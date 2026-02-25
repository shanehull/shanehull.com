FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community hugo

RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates

RUN go install github.com/a-h/templ/cmd/templ@latest

RUN HUGO_ENV=production hugo --cleanDestinationDir

RUN templ generate

RUN go build -o bin/main ./cmd/server/

FROM scratch

ENV ALLOWED_ORIGIN="*"
ENV SERVER_HOST="127.0.0.1"

COPY --from=builder /app/bin/main ./bin/main
COPY --from=builder /app/public ./static/public

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 1314

CMD ["./bin/main"]
