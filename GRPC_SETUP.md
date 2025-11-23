# gRPC Setup và Implementation Guide

## ✅ Đã hoàn thành

### 1. Proto Files

- ✅ `shared/proto/auth/auth.proto`
- ✅ `shared/proto/logger/logger.proto`
- ✅ `shared/proto/mail/mail.proto`
- ✅ `shared/proto/image/image.proto`

### 2. Go Services - gRPC Servers

- ✅ Logger Service - gRPC server trên port 50001
- ✅ Mail Service - gRPC server trên port 50002
- ✅ Image Service - gRPC server trên port 50003
- ✅ Auth Service - gRPC server trên port 50000 (đã implement đầy đủ handlers)

### 3. .NET Gateway - gRPC Clients

- ✅ Thêm gRPC packages vào RideSharing.Api
- ✅ Cấu hình proto files để generate C# code
- ✅ Tạo GrpcClients service
- ✅ Tạo HTTP endpoints để gọi gRPC:
  - Auth endpoints (SignUp, SignIn, VerifyEmail, RefreshToken, GetUser, GetCities)
  - Logger endpoints
  - Mail endpoints
  - Image endpoints

## 📋 Cần làm tiếp

### 1. Generate Proto Code

#### Go Services:

**Lưu ý**: Cần cài đặt `protoc` và các plugins trước:

```bash
# Install protoc (macOS)
brew install protobuf

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Sau đó generate code:

```bash
cd shared/proto
make generate
# hoặc
./generate.sh
```

#### .NET Gateway:

Code sẽ tự động generate khi build project (nhờ Grpc.Tools).

### 2. Cập nhật go.mod

Đã thêm replace directive vào `services/auth/go.mod`:

```go
replace ride-sharing/shared/generated => ../../shared/generated
```

Các services khác (logger, mail, image) cũng cần thêm replace directive tương tự nếu chưa có.

## 🔧 Cấu hình

### Ports

| Service | gRPC Port | HTTP Port | K8s Service          |
| ------- | --------- | --------- | -------------------- |
| Auth    | 50000     | 80 (8080) | auth:50000           |
| Logger  | 50001     | 80 (8082) | logger-service:50001 |
| Mail    | 50002     | 80 (8083) | mail-service:50002   |
| Image   | 50003     | 80 (8085) | image-service:50003  |

### appsettings.json

```json
{
  "Grpc": {
    "AuthService": { "Url": "http://auth:50000" },
    "LoggerService": { "Url": "http://logger-service:50001" },
    "MailService": { "Url": "http://mail-service:50002" },
    "ImageService": { "Url": "http://image-service:50003" }
  }
}
```

## 🚀 Cách sử dụng

### Frontend vẫn gọi HTTP:

```typescript
// Frontend không thay đổi
POST http://localhost:8084/api/v1/Auth/sign-up
POST http://localhost:8084/api/v1/Auth/sign-in
POST http://localhost:8084/api/v1/Auth/verify-email
POST http://localhost:8084/api/v1/Auth/refresh-token
GET  http://localhost:8084/api/v1/Auth/user/{userId}
GET  http://localhost:8084/api/v1/Auth/cities
POST http://localhost:8084/api/v1/Logger/log
POST http://localhost:8084/api/v1/Mail/send
POST http://localhost:8084/api/v1/Image/upload/folder
```

### Gateway tự động convert HTTP → gRPC:

```
Frontend (HTTP)
  → API Gateway (HTTP endpoint)
    → gRPC Client
      → Go Service (gRPC server)
        → Response
      ← gRPC Response
    ← Convert to HTTP
  ← HTTP Response
```

## 📝 Notes

- Các services vẫn chạy HTTP server song song (backward compatibility)
- YARP reverse proxy vẫn được giữ lại (có thể xóa sau khi test xong)
- Frontend không cần thay đổi gì
- Tất cả giao tiếp giữa gateway và services giờ dùng gRPC
