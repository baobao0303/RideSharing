package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const webPort = "80"

func main() {
	// Ensure uploads directory exists
	uploadDir := getUploadRoot()
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	// Start gRPC server
	go startGRPCServer(uploadDir)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// Serve uploaded files
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", webPort),
		Handler:          mux,
		ReadTimeout:      15 * time.Second,
		WriteTimeout:     30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Println("image-service HTTP listening on :" + webPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}

func getUploadRoot() string {
	// Prefer /app/uploads in container, fallback to ./uploads locally
	candidates := []string{
		"/app/uploads",
		filepath.Join(".", "uploads"),
	}
	for _, p := range candidates {
		// first writable dir will be used
		if err := os.MkdirAll(p, 0755); err == nil {
			return p
		}
	}
	return filepath.Join(".", "uploads")
}

