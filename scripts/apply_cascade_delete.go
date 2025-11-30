package main

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	dbType := os.Getenv("DB_TYPE")
	dsn := os.Getenv("DB_DSN")

	if dbType != "sqlite" {
		log.Fatal("This script is for SQLite only. For MySQL, run the SQL migration manually.")
	}

	if dsn == "" {
		dsn = "dmIntegroff.db"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Applying CASCADE DELETE migration for SQLite...")

	// SQLite doesn't support ALTER TABLE to modify constraints
	// We need to recreate the table
	sqlStatements := []string{
		"PRAGMA foreign_keys=off;",
		"BEGIN TRANSACTION;",
		
		// Create new table with CASCADE
		`CREATE TABLE integrations_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT NOT NULL,
			webhook_token TEXT UNIQUE NOT NULL,
			target_api TEXT,
			mode TEXT DEFAULT 'inactive',
			sample_payload TEXT,
			mapping_config TEXT,
			project_id INTEGER NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`,
		
		// Copy data
		`INSERT INTO integrations_new 
		SELECT id, created_at, updated_at, deleted_at, name, webhook_token, 
		       target_api, mode, sample_payload, mapping_config, project_id
		FROM integrations;`,
		
		// Drop old table
		"DROP TABLE integrations;",
		
		// Rename new table
		"ALTER TABLE integrations_new RENAME TO integrations;",
		
		// Recreate index
		"CREATE INDEX idx_integrations_deleted_at ON integrations(deleted_at);",
		"CREATE INDEX idx_integrations_webhook_token ON integrations(webhook_token);",
		
		"COMMIT;",
		"PRAGMA foreign_keys=on;",
	}

	for _, sql := range sqlStatements {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Error executing: %s\nError: %v", sql, err)
			db.Exec("ROLLBACK;")
			log.Fatal("Migration failed")
		}
	}

	log.Println("✅ CASCADE DELETE migration applied successfully!")
}
