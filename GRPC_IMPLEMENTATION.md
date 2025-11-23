# gRPC Implementation Guide

## Tổng quan

Project đã được chuyển đổi để sử dụng gRPC cho giao tiếp giữa các services.

## Cấu trúc

### Proto Files
- `shared/proto/auth/auth.proto` - Auth service definitions
- `shared/proto/logger/logger.proto` - Logger service definitions
- `shared/proto/mail/mail.proto` - Mail service definitions
- `shared/proto/image/image.proto` - Image service definitions

### Generated Code
- `shared/generated/` - Go code được generate từ proto files

## Ports

| Service | gRPC Port | HTTP Port |
|---------|-----------|-----------|
| Auth | 50000 | 80 (8080) |
| Logger | 50001 | 80 (8082) |
| Mail | 50002 | 80 (8083) |
| Image | 50003 | 80 (8085) |

## Generate Code

### Prerequisites

```bash
# Install protoc
brew install protobuf  # macOS
# hoặc
sudo apt-get install protobuf-compiler  # Linux

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate

```bash
cd shared/proto
make generate
# hoặc
./generate.sh
```

## Implementation Status

### ✅ Completed
- [x] Proto files cho tất cả services
- [x] gRPC server cho Logger Service
- [x] gRPC server cho Mail Service
- [x] gRPC server cho Image Service
- [x] gRPC server structure cho Auth Service

### ⚠️ TODO
- [ ] Implement đầy đủ Auth Service gRPC handlers
- [ ] Implement gRPC client trong .NET api-net-gateway
- [ ] Cập nhật gateway để convert HTTP → gRPC
- [ ] Testing và validation

## Next Steps

1. **Complete Auth Service**: Implement đầy đủ các handlers trong `grpc_server.go`
2. **.NET Gateway**: Implement gRPC clients trong api-net-gateway
3. **HTTP to gRPC Bridge**: Tạo middleware để convert HTTP requests thành gRPC calls
4. **Testing**: Test end-to-end flow

## Notes

- Các services vẫn chạy HTTP server song song với gRPC server (backward compatibility)
- Legacy RPC (net/rpc) vẫn được giữ lại
- Frontend vẫn gọi HTTP → Gateway sẽ convert sang gRPC

