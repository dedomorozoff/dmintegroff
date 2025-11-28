package database

import (
	"dmintegroff/internal/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	log.Println("Default admin user created (admin/admin)")
}
