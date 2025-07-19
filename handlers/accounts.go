package handlers

import (
	"bank-app/config"
	"bank-app/models"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GenerateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(1000000000))
}

func CreateAccount(c *gin.Context) {
	// Extract data from the request body
	var accountRequest struct {
		UserID         uint    `json:"user_id"`
		AccountType    string  `json:"account_type"`
		InitialBalance float64 `json:"initial_balance"`
	}

	// Bind JSON data to accountRequest struct
	if err := c.ShouldBindJSON(&accountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if the user exists
	var user models.User
	if err := config.DB.First(&user, accountRequest.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		}
		return
	}

	// Create a new account for the existing user
	account := models.Account{
		UserID:      accountRequest.UserID,
		AccountType: accountRequest.AccountType,
		Balance:     accountRequest.InitialBalance,
		AccountNo:   GenerateAccountNumber(), // Generate account number
	}

	// Save the new account to the database
	if err := config.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	// Return the created account details
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Account created successfully",
		"account_id": account.ID,
		"account_no": account.AccountNo,
		"balance":    account.Balance,
	})
}

func DeleteAccount(c *gin.Context) {
	accountNo := c.Param("account_no")
	var account models.Account

	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find account"})
		}
		return
	}

	if err := config.DB.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func Deposit(c *gin.Context) {
	accountNo := c.Param("account_no")
	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Find the account
	var account models.Account
	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Add the deposit amount to the balance
	account.Balance += request.Amount
	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// Log the transaction
	transaction := models.Transaction{
		TransactionType: "deposit",
		Amount:          request.Amount,
		AccountID:       account.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deposit successful",
		"balance": account.Balance,
	})
}

func Withdraw(c *gin.Context) {
	accountNo := c.Param("account_no")
	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Find the account
	var account models.Account
	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Check if the account has enough balance
	if account.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Deduct the amount from the balance
	account.Balance -= request.Amount
	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// Log the transaction
	transaction := models.Transaction{
		TransactionType: "withdrawal",
		Amount:          request.Amount,
		AccountID:       account.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Withdrawal successful",
		"balance": account.Balance,
	})
}

func Transfer(c *gin.Context) {
	fromAccountNo := c.Param("from_account")
	toAccountNo := c.Param("to_account")
	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Find the sender's account
	var fromAccount models.Account
	if err := config.DB.Where("account_no = ?", fromAccountNo).First(&fromAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender account not found"})
		return
	}

	// Find the receiver's account
	var toAccount models.Account
	if err := config.DB.Where("account_no = ?", toAccountNo).First(&toAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver account not found"})
		return
	}

	// Check if the sender has enough balance
	if fromAccount.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Deduct from sender and add to receiver
	fromAccount.Balance -= request.Amount
	toAccount.Balance += request.Amount

	if err := config.DB.Save(&fromAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sender account"})
		return
	}

	if err := config.DB.Save(&toAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receiver account"})
		return
	}

	// Log the transaction for the sender
	transactionFrom := models.Transaction{
		TransactionType: "transfer",
		Amount:          request.Amount,
		AccountID:       fromAccount.ID,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     &toAccount.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := config.DB.Create(&transactionFrom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log sender transaction"})
		return
	}

	// Log the transaction for the receiver
	transactionTo := models.Transaction{
		TransactionType: "transfer",
		Amount:          request.Amount,
		AccountID:       toAccount.ID,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     &toAccount.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := config.DB.Create(&transactionTo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log receiver transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer successful",
		"balance": fromAccount.Balance,
	})
}
