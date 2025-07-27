package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/infrastructure/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create migrations table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}

	// Get migration files
	migrationDir := "internal/infrastructure/database/migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		log.Fatal("Failed to read migrations directory:", err)
	}

	// Sort migration files
	var migrationFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute migrations
	for _, filename := range migrationFiles {
		// Check if migration already executed
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM migrations WHERE filename = $1", filename)
		if err != nil {
			log.Printf("Error checking migration %s: %v", filename, err)
			continue
		}

		if count > 0 {
			log.Printf("Migration %s already executed, skipping...", filename)
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationDir, filename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read migration file %s: %v", filename, err)
			continue
		}

		// Execute migration
		log.Printf("Executing migration: %s", filename)
		_, err = db.Exec(string(content))
		if err != nil {
			log.Printf("Failed to execute migration %s: %v", filename, err)
			continue
		}

		// Record migration as executed
		_, err = db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
		if err != nil {
			log.Printf("Failed to record migration %s: %v", filename, err)
			continue
		}

		log.Printf("Migration %s executed successfully", filename)
	}

	log.Println("All migrations completed!")
}
