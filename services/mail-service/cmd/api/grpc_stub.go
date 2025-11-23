//go:build !grpc

package main

// startGRPCServer is a no-op when built without the 'grpc' build tag.
func startGRPCServer(_ Mail) {
	// gRPC disabled for this build
}


