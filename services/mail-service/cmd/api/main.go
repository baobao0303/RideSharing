package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Mailer Mail
}

const (
	webPort = "80"
	rpcPort = "5002"
)

func main() {
	// Load .env file from project root
	loadEnvFile()

	app := Config{
		Mailer: createMail(),
	}

	// Start RPC server in background
	go app.rpcListen()
	// Start gRPC server in background
	startGRPCServer(app.Mailer)

	log.Println("Starting mail service on port", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func createMail() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
	}

	return m
}

// loadEnvFile loads .env file from project root
func loadEnvFile() {
	var envPath string
	var found bool

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("⚠️  Could not get working directory: %v", err)
		cwd = "."
	}

	// List of possible .env file locations to try
	possiblePaths := []string{
		".env",                                    // Current directory
		filepath.Join("..", ".env"),              // Parent directory
		filepath.Join("../..", ".env"),           // 2 levels up
		filepath.Join("../../..", ".env"),        // 3 levels up
		filepath.Join(cwd, ".env"),               // Absolute from current dir
		filepath.Join(cwd, "..", ".env"),         // Absolute parent
		filepath.Join(cwd, "../..", ".env"),      // Absolute 2 levels up
		"/app/.env",                              // Common container path
		"/app/../.env",                           // Container parent
	}

	// Try each path
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			envPath = path
			found = true
			break
		}
	}

	// If not found, try searching up from current directory
	if !found {
		dir := cwd
		for i := 0; i < 10; i++ {
			testPath := filepath.Join(dir, ".env")
			if _, err := os.Stat(testPath); err == nil {
				envPath = testPath
				found = true
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break // Reached root
			}
			dir = parent
		}
	}

	if found {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("⚠️  Found .env at %s but failed to load: %v", envPath, err)
		} else {
			log.Printf("✅ Loaded environment variables from: %s", envPath)
		}
	} else {
		log.Printf("⚠️  .env file not found (searched from: %s), using environment variables or defaults", cwd)
	}
}

// rpcListen starts the RPC server
func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port", rpcPort)

	// Register RPCServer with Mailer
	rpcServer := &RPCServer{
		Mailer: app.Mailer,
	}
	err := rpc.Register(rpcServer)
	if err != nil {
		log.Printf("Error registering RPC server: %v", err)
		return err
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	log.Println("RPC server listening on port", rpcPort)

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			log.Printf("Error accepting RPC connection: %v", err)
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}
