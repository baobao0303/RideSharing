//go:build grpc
// +build grpc

package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	loggerpb "ride-sharing/shared/generated/logger"
	"log-service/data"
)

const grpcPort = 50001

type loggerGrpcServer struct {
	loggerpb.UnimplementedLoggerServiceServer
	models data.Models
}

func startGRPCServer(models data.Models) {
	addr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("logger-service failed to listen: %v", err)
	}

	s := grpc.NewServer()
	loggerpb.RegisterLoggerServiceServer(s, &loggerGrpcServer{models: models})
	reflection.Register(s)

	log.Printf("logger-service gRPC listening on %s", addr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
}

func (s *loggerGrpcServer) LogInfo(ctx context.Context, req *loggerpb.LogRequest) (*loggerpb.LogResponse, error) {
	e := data.LogEntry{Name: req.GetName(), Data: req.GetData()}
	if err := s.models.LogEntry.Insert(e); err != nil {
		return nil, err
	}
	return &loggerpb.LogResponse{Message: "logged"}, nil
}

