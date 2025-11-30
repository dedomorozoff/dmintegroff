package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default/env vars")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "dmIntegroff.db"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Fixing missing webhook_token values...")

	// Check if webhook_token column exists
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM pragma_table_info('integrations') WHERE name='webhook_token'").Scan(&count).Error
	if err != nil {
		log.Fatal("Failed to check column:", err)
	}

	if count == 0 {
		log.Println("webhook_token column doesn't exist yet, nothing to fix")
		return
	}

	// Get integrations without webhook_token
	type Integration struct {
		ID           uint
		WebhookToken string
	}

	var integrations []Integration
	err = db.Raw("SELECT id, webhook_token FROM integrations WHERE webhook_token IS NULL OR webhook_token = ''").Scan(&integrations).Error
	if err != nil {
		log.Fatal("Failed to query integrations:", err)
	}

	if len(integrations) == 0 {
		log.Println("✅ All integrations already have webhook_token")
		return
	}

	log.Printf("Found %d integrations without webhook_token, generating...", len(integrations))

	for _, integration := range integrations {
		token := generateToken()
		err = db.Exec("UPDATE integrations SET webhook_token = ? WHERE id = ?", token, integration.ID).Error
		if err != nil {
			log.Printf("Failed to update integration %d: %v", integration.ID, err)
		} else {
			log.Printf("✅ Generated token for integration %d", integration.ID)
		}
	}

	log.Println("✅ All webhook tokens fixed!")
}
