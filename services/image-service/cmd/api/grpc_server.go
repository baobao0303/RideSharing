package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	imagepb "ride-sharing/shared/generated/image"
)

const grpcPort = 50003

type imageGrpcServer struct {
	imagepb.UnimplementedImageServiceServer
	uploadRoot string
}

func startGRPCServer(uploadRoot string) {
	addr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("image-service failed to listen: %v", err)
	}

	s := grpc.NewServer()
	imagepb.RegisterImageServiceServer(s, &imageGrpcServer{uploadRoot: uploadRoot})
	reflection.Register(s)

	log.Printf("image-service gRPC listening on %s", addr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
}

func (s *imageGrpcServer) UploadToFolder(ctx context.Context, req *imagepb.UploadRequest) (*imagepb.UploadResponse, error) {
	folder := sanitize(req.GetFolder())
	fileName := sanitize(req.GetFileName())
	if folder == "" || fileName == "" {
		return nil, fmt.Errorf("folder and fileName are required")
	}

	dir := filepath.Join(s.uploadRoot, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	fullPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(fullPath, req.GetContent(), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	urlPath := "/uploads/" + folder + "/" + fileName
	return &imagepb.UploadResponse{Url: urlPath, Path: fullPath, Message: "uploaded"}, nil
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/\\")
	s = strings.ReplaceAll(s, "..", "")
	return s
}


