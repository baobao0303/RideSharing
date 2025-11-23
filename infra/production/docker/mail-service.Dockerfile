# Multi-stage build for mail-service (build inside container)
FROM golang:1.23 AS builder
WORKDIR /app
# Copy only the mail-service module
COPY services/mail-service/ ./
# Prepare module proxy and tidy dependencies (generates go.sum)
ENV GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org
RUN go mod tidy
# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mail-service ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/mail-service ./mail-service
ENTRYPOINT ["./mail-service"]

