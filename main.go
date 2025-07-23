package main

import (
	"bank-app/config"
	"bank-app/handlers"
	"bank-app/middleware"
	"bank-app/rabbitmq"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

//	func init() {
//		// If running locally, you may still want to load the .env file.
//		// Uncomment the following lines if you want to keep godotenv loading for local development.
//		if os.Getenv("ENV") != "production" {
//			err := godotenv.Load()
//			if err != nil {
//				log.Fatal("Error loading .env file")
//			}
//		}
//	}
func main() {
	// Connect to the database
	config.ConnectDB()
	defer config.CloseDB()

	if err := rabbitmq.Init(); err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// Set up the Gin router
	r := gin.Default()

	// Add this before defining routes in `main.go`
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"}, // or use env var
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	AllowCredentials: true,
	// }))

	// Dummy base URL endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the Bank API",
		})
	})

	r.POST("/signup", handlers.SignUp)
	r.POST("/login", handlers.Login)
	r.GET("/transactions/summary", handlers.GetAllTransactionsSummary)

	auth := r.Group("/")
	auth.Use(middleware.JWTAuthMiddleware())

	// Routes for users
	// r.POST("/users", handlers.CreateUser)
	auth.GET("/users/:id", handlers.GetUserByID)
	auth.GET("/accounts", handlers.GetAllAccounts)

	// Routes for accounts
	auth.POST("/accounts", handlers.CreateAccount)
	auth.POST("/accounts/:account_no/deposit", handlers.Deposit)
	auth.POST("/accounts/:account_no/withdraw", handlers.Withdraw)

	// Update the transfer route to avoid conflict
	auth.POST("/accounts/transfer/:from_account/:to_account", handlers.Transfer)

	auth.GET("/transactions/:id", handlers.GetTransactionByID)
	auth.GET("/users/:id/transactions", handlers.GetTransactionsByUserID)
	auth.GET("/accounts/:account_no/transactions", handlers.GetTransactionsByAccountNo)

	// Start the server on the port from the environment or default to 7070
	port := "8080"

	// Start the Gin server
	err := r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
