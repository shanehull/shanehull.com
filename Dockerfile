FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . .

RUN apk update && apk upgrade && apk add --no-cache ca-certificates hugo
RUN update-ca-certificates

RUN go install github.com/a-h/templ/cmd/templ@latest

RUN GO_ENABLED=0 GOOS=linux go run cmd/build/main.go

FROM scratch

ENV SH_SERVER_HOST=""

COPY --from=builder /app/build/server ./build/server
COPY --from=builder /app/build/pages.gob ./build/pages.gob
COPY --from=builder /app/public ./public

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["./build/server"]
