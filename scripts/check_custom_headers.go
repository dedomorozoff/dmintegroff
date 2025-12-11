package main

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"fmt"
	"log"
)

func main() {
	// Initialize database
	database.Connect()

	var integrations []models.Integration
	if err := database.DB.Find(&integrations).Error; err != nil {
		log.Fatal("Failed to query integrations:", err)
	}

	fmt.Printf("Found %d integrations:\n\n", len(integrations))
	
	for _, integration := range integrations {
		fmt.Printf("ID: %d\n", integration.ID)
		fmt.Printf("Name: %s\n", integration.Name)
		fmt.Printf("Custom Headers: %s\n", integration.CustomHeaders)
		fmt.Printf("---\n\n")
	}
}
