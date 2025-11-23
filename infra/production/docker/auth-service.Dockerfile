# Multi-stage build for auth-service (build inside container with full repo context)
FROM golang:1.23 AS builder
WORKDIR /workspace
COPY . .
WORKDIR /workspace/services/auth
ENV GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/auth ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/auth ./auth
ENTRYPOINT ["./auth"]

