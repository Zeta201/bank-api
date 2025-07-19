package main

import (
	"bank-app/config"
	"bank-app/handlers"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Import godotenv to load environment variables
)

func init() {
	// Load the .env file at the start
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Connect to the database
	config.ConnectDB()
	defer config.CloseDB()

	// Set up the Gin router
	r := gin.Default()

	// Routes for users
	r.POST("/users", handlers.CreateUser)
	r.GET("/users/:id", handlers.GetUserByID)

	// Routes for accounts
	r.POST("/accounts", handlers.CreateAccount)
	r.DELETE("/accounts/:account_no", handlers.DeleteAccount)
	r.POST("/accounts/:account_no/deposit", handlers.Deposit)
	r.POST("/accounts/:account_no/withdraw", handlers.Withdraw)

	// Update the transfer route to avoid conflict
	r.POST("/accounts/transfer/:from_account/:to_account", handlers.Transfer)

	r.GET("/transactions", handlers.GetAllTransactions)
	r.GET("/transactions/:id", handlers.GetTransactionByID)
	r.GET("/users/:id/transactions", handlers.GetTransactionsByUserID)
	r.GET("/accounts/:account_no/transactions", handlers.GetTransactionsByAccountNo)

	// Start the server
	port := "7070" // Default port if not set in .env

	r.Run(fmt.Sprintf(":%s", port))
}
