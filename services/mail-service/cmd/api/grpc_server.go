//go:build grpc

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	mailpb "ride-sharing/shared/generated/mail"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const grpcPort = 50002

type mailGrpcServer struct {
	mailpb.UnimplementedMailServiceServer
	mailer Mail
}

func startGRPCServer(m Mail) {
	addr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("mail-service failed to listen: %v", err)
	}

	s := grpc.NewServer()
	mailpb.RegisterMailServiceServer(s, &mailGrpcServer{mailer: m})
	reflection.Register(s)

	log.Printf("mail-service gRPC listening on %s", addr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
}

func (s *mailGrpcServer) SendMail(ctx context.Context, req *mailpb.MailRequest) (*mailpb.MailResponse, error) {
	msg := Message{
		From:        req.GetFrom(),
		FromName:    req.GetFromName(),
		To:          req.GetTo(),
		Subject:     req.GetSubject(),
		Data:        req.GetMessage(),
		Attachments: req.GetAttachments(),
	}
	if msg.From == "" { msg.From = s.mailer.FromAddress }
	if msg.FromName == "" { msg.FromName = s.mailer.FromName }
	if err := s.mailer.SendSMTPMessage(msg); err != nil {
		return nil, err
	}
	return &mailpb.MailResponse{Message: "Email sent successfully to " + req.GetTo()}, nil
}

// helper to construct Mail from env (if needed elsewhere)
func buildMailFromEnv() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	return Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
	}
}

