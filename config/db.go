package config

import (
	"bank-app/models"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	// No need to load .env file in Choreo, environment variables are passed through the platform
	// If you're still in local dev and using `.env`, you could keep godotenv in a conditional block.
	// Uncomment the following lines if running locally (comment it out for production):
	// if os.Getenv("ENV") != "production" {
	// 	err = godotenv.Load()
	// 	if err != nil {
	// 		log.Fatal("Error loading .env file")
	// 	}
	// }

	// Retrieve environment variables directly
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	// Open the database connection
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	log.Println("Database connected successfully!")

	// Automatically migrate the schema
	err = DB.AutoMigrate(
		&models.User{},    // Migrating the User model
		&models.Account{}, // Migrating the Account model
		&models.Transaction{},
	).Error
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("Database schema migrated successfully!")
}

func CloseDB() {
	if err := DB.Close(); err != nil {
		log.Fatal("Failed to close the database connection: ", err)
	}
}
