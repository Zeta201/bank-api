package handlers

import (
	"bank-app/config"
	"bank-app/models"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateUniqueAccountNumber() string {
	for {
		accountNo := fmt.Sprintf("%09d", rand.Intn(1_000_000_000))
		var existing models.Account
		if config.DB.Where("account_no = ?", accountNo).First(&existing).RecordNotFound() {
			return accountNo
		}
	}
}

func CreateAccount(c *gin.Context) {
	var accountRequest models.AccountRequest

	if err := c.ShouldBindJSON(&accountRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid input"})
		return
	}

	if accountRequest.InitialBalance < 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Initial balance cannot be negative"})
		return
	}

	validTypes := map[string]bool{"savings": true, "checking": true}
	if !validTypes[accountRequest.AccountType] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid account type"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "User not found"})
		return
	}

	// Check if this account type already exists for the user
	var existing models.Account
	if err := config.DB.Where("user_id = ? AND account_type = ?", userID, accountRequest.AccountType).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Account type already exists"})
		return
	}

	account := models.Account{
		UserID:      userID,
		AccountType: accountRequest.AccountType,
		Balance:     accountRequest.InitialBalance,
		AccountNo:   GenerateUniqueAccountNumber(),
	}

	if err := config.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Account created successfully",
		"account_id":   account.ID,
		"account_no":   account.AccountNo,
		"balance":      account.Balance,
		"account_type": account.AccountType,
	})
}

// func DeleteAccount(c *gin.Context) {
// 	accountNo := c.Param("account_no")
// 	var account models.Account

// 	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find account"})
// 		}
// 		return
// 	}

// 	if err := config.DB.Delete(&account).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
// }

func Deposit(c *gin.Context) {
	accountNo := c.Param("account_no")

	var request models.AmountRequest

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

	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing amount"})
		return
	}

	// Get authenticated user's ID from context
	userID := c.MustGet("userID").(uint)

	// Find the account AND ensure it belongs to the authenticated user
	var account models.Account
	if err := config.DB.Where("account_no = ? AND user_id = ?", accountNo, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account not found or access denied"})
		return
	}

	// Check if the account has enough balance
	if account.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Deduct the amount
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

	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing amount"})
		return
	}

	userID := c.MustGet("userID").(uint)

	// Ensure the 'from' account belongs to the logged-in user
	var fromAccount models.Account
	if err := config.DB.Where("account_no = ? AND user_id = ?", fromAccountNo, userID).First(&fromAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this account"})
		return
	}

	// Lookup receiver's account
	var toAccount models.Account
	if err := config.DB.Where("account_no = ?", toAccountNo).First(&toAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver account not found"})
		return
	}

	// Prevent transferring to the same account
	if fromAccount.AccountNo == toAccount.AccountNo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	// Check sufficient balance
	if fromAccount.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Execute transfer
	fromAccount.Balance -= request.Amount
	toAccount.Balance += request.Amount

	// Use transaction to ensure atomicity
	tx := config.DB.Begin()
	if err := tx.Save(&fromAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sender account"})
		return
	}
	if err := tx.Save(&toAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receiver account"})
		return
	}

	// Log sender transaction
	transactionFrom := models.Transaction{
		TransactionType: "transfer",
		Amount:          request.Amount,
		AccountID:       fromAccount.ID,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     &toAccount.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transactionFrom).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log sender transaction"})
		return
	}

	// Log receiver transaction
	transactionTo := models.Transaction{
		TransactionType: "transfer",
		Amount:          request.Amount,
		AccountID:       toAccount.ID,
		FromAccountID:   &fromAccount.ID,
		ToAccountID:     &toAccount.ID,
		Status:          "success",
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transactionTo).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log receiver transaction"})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer successful",
		"balance": fromAccount.Balance,
	})
}

func GetAllAccounts(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var accounts []models.Account
	if err := config.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}
