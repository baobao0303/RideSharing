# Multi-stage build for logger-service (build inside container)
FROM golang:1.23 AS builder
WORKDIR /workspace
# Copy full repo so replace ../../shared/generated works
COPY . .
WORKDIR /workspace/services/logger-service
# Prepare module proxy and tidy dependencies (generates go.sum)
ENV GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org
RUN go mod tidy
# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/logger-service ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/logger-service ./logger-service
ENTRYPOINT ["./logger-service"]

