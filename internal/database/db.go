package database

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	dbType := os.Getenv("DB_TYPE") // "sqlite" or "mysql"
	dsn := os.Getenv("DB_DSN")     // Connection string

	if dbType == "mysql" {
		if dsn == "" {
			dsn = "user:password@tcp(127.0.0.1:3306)/dmIntegroff?charset=utf8mb4&parseTime=True&loc=Local"
		}
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		// Default to SQLite
		if dsn == "" {
			dsn = "dmIntegroff.db"
		}
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}

func Migrate(dst ...interface{}) {
	if err := DB.AutoMigrate(dst...); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")
}
