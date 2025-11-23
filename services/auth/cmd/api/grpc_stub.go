//go:build !grpc

package main

import (
	"ride-sharing/services/auth/data"
)

// startGRPCServer is a no-op when built without the 'grpc' build tag.
func startGRPCServer(models *data.Models) {
	// gRPC disabled for this build
}


