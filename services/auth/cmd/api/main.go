package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"ride-sharing/services/auth/data"
	"strings"
	"time"

	_ "ride-sharing/services/auth/cmd/api/docs" // Swagger docs
	"ride-sharing/services/auth/cmd/api/handlers"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

// @title Auth Service API
// @version 1.0
// @description Authentication Service API Documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @host localhost:8080
// @BasePath /

const webPort = "80"
var counts int64

type Config struct {
	DB *sql.DB
	Models data.Models
}

func main() {
	// Load .env file from project root
	loadEnvFile()

	//log the start of the application
	log.Println("Starting authentication service")


	// connect to database
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %s\n", err)
		os.Exit(1)
	}

	// run migrations
	if err := runMigrations(db); err != nil {
		log.Printf("Warning: Error running migrations: %v\n", err)
		// Continue anyway - migrations might already be applied
	}

	// setup config
	app := Config{
		DB: db,
		Models: data.New(db),
	}

	// Set models instance for handlers
	handlers.SetModels(&app.Models)

	// Start gRPC server in background
	startGRPCServer(&app.Models)

	src := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = src.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}


//configure database
func openDB(dsn string) (*sql.DB, error) {
	// if using env os.Getenv("DB_URL")
	db, err := sql.Open("pgx", dsn )
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() (*sql.DB, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=password dbname=users sslmode=disable"
	}
	log.Println("Attempting to connect to PostgreSQL...")
	for{
		db, err := openDB(dsn)
		if err != nil {
			log.Printf("Postgres not ready yet... (attempt %d/10)", counts+1)
			counts++
			if counts > 10 {
				log.Fatalf("Failed to connect to PostgreSQL after 10 attempts: %v", err)
				return nil, err
			}
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("✅ Successfully connected to PostgreSQL!")
		return db, nil
	}
}

// runMigrations executes all SQL migration files in the migrations directory
func runMigrations(db *sql.DB) error {
	// Try multiple possible paths for migrations
	migrationPaths := []string{
		"./migrations",
		"/app/migrations",
		"../../migrations",
		"../migrations",
		"services/auth/migrations",
	}

	var migrationsDir string
	for _, path := range migrationPaths {
		if _, err := os.Stat(path); err == nil {
			migrationsDir = path
			break
		}
	}

	if migrationsDir == "" {
		log.Println("⚠️  Migrations directory not found, skipping migrations")
		return nil
	}

	log.Printf("📦 Running migrations from: %s\n", migrationsDir)

	// Read all SQL files in migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %v", err)
	}

	// Sort files by name to ensure correct order
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	if len(sqlFiles) == 0 {
		log.Println("⚠️  No migration files found")
		return nil
	}

	// Execute each migration file
	for _, fileName := range sqlFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		log.Printf("🔄 Running migration: %s\n", fileName)

		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("❌ Error reading migration file %s: %v\n", fileName, err)
			continue
		}

		// Execute the SQL
		_, err = db.Exec(string(sqlContent))
		if err != nil {
			// Check if error is because table already exists (migration already run)
			if strings.Contains(err.Error(), "already exists") || 
			   strings.Contains(err.Error(), "duplicate key") {
				log.Printf("ℹ️  Migration %s already applied, skipping\n", fileName)
				continue
			}
			return fmt.Errorf("error executing migration %s: %v", fileName, err)
		}

		log.Printf("✅ Migration %s completed successfully\n", fileName)
	}

	log.Println("✅ All migrations completed!")
	return nil
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