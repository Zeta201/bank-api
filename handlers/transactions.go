package handlers

import (
	"bank-app/config"
	"bank-app/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	if err := config.DB.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func GetTransactionByID(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction
	if err := config.DB.First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		}
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func GetTransactionsByUserID(c *gin.Context) {
	userID := c.Param("id")
	var accounts []models.Account

	// Fetch all accounts for the given user
	if err := config.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Fetch transactions for the user's accounts
	var transactions []models.Transaction
	if err := config.DB.Where("account_id IN (?)", getAccountIDs(accounts)).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// Helper function to extract account IDs from accounts slice
func getAccountIDs(accounts []models.Account) []uint {
	var accountIDs []uint
	for _, account := range accounts {
		accountIDs = append(accountIDs, account.ID)
	}
	return accountIDs
}

func GetTransactionsByAccountNo(c *gin.Context) {
	accountNo := c.Param("account_no")
	var account models.Account

	// Fetch the account by account number
	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Fetch transactions for the given account
	var transactions []models.Transaction
	if err := config.DB.Where("account_id = ?", account.ID).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
