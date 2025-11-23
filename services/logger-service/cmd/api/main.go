package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// Load .env file from project root
	loadEnvFile()

	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	go app.rpcListen()
	// start gRPC server
	startGRPCServer(app.Models)
	
	// start web server
	log.Println("Starting service on port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}

}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	
	// Register RPCServer with Models
	rpcServer := &RPCServer{
		Models: app.Models,
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

func connectToMongo() (*mongo.Client, error) {
	// Get MongoDB URL from environment variable
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:password@mongo:27017"
	}

	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	
	// Only set auth if URL doesn't already contain credentials
	if !strings.Contains(mongoURL, "@") {
		username := os.Getenv("MONGO_USERNAME")
		if username == "" {
			username = "admin"
		}
		password := os.Getenv("MONGO_PASSWORD")
		if password == "" {
			password = "password"
		}
		
		clientOptions.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})
	}

	// connect
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	log.Println("Connected to mongo!")

	return c, nil
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
