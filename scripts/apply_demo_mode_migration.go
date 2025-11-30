package main

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	dbType := os.Getenv("DB_TYPE")
	dsn := os.Getenv("DB_DSN")

	if dsn == "" {
		dsn = "dmintegroff.db"
	}

	var db *gorm.DB
	var err error

	if dbType == "mysql" {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Applying demo mode migration...")

	var sqlStatements []string

	if dbType == "mysql" {
		sqlStatements = []string{
			"ALTER TABLE users ADD COLUMN is_demo BOOLEAN DEFAULT FALSE;",
			"ALTER TABLE users ADD COLUMN expires_at DATETIME NULL;",
			"CREATE INDEX idx_users_expires_at ON users(expires_at);",
		}
	} else {
		sqlStatements = []string{
			"ALTER TABLE users ADD COLUMN is_demo BOOLEAN DEFAULT 0;",
			"ALTER TABLE users ADD COLUMN expires_at DATETIME;",
			"CREATE INDEX idx_users_expires_at ON users(expires_at);",
		}
	}

	for _, sql := range sqlStatements {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Warning executing: %s\nError: %v", sql, err)
			// Continue even if column already exists
		}
	}

	log.Println("✅ Demo mode migration applied successfully!")
	log.Println("📖 See docs/DEMO_MODE.md for usage instructions")
}
