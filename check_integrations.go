package main

import (
	"fmt"
	"log"
	
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Integration struct {
	ID   uint
	Name string
}

func main() {
	godotenv.Load()
	
	db, err := gorm.Open(sqlite.Open("dmIntegroff.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	
	var integrations []Integration
	db.Table("integrations").Select("id, name").Limit(10).Find(&integrations)
	
	fmt.Println("Интеграции в базе:")
	for _, i := range integrations {
		fmt.Printf("ID: %d, Name: %s\n", i.ID, i.Name)
	}
}
