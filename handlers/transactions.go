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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// @Summary      Get transaction by ID
// @Description  Retrieve transaction details by transaction ID
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        id   path      string             true  "Transaction ID"
// @Success      200  {object}  models.Transaction
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /transactions/{id} [get]
func GetTransactionByID(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction
	if err := config.DB.First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to retrieve transaction"})
		}
		return
	}
	c.JSON(http.StatusOK, transaction)
}

// @Summary      Get transactions by user ID
// @Description  Retrieves all transactions for all accounts belonging to a specific user.
// @Tags         transactions
// @Param        id   path      string  true  "User ID"
// @Produce      json
// @Success      200  {array}   models.Transaction
// @Failure      404  {object}  models.ErrorResponse  "User not found"
// @Failure      500  {object}  models.ErrorResponse  "Failed to fetch transactions"
// @Router       /users/{id}/transactions [get]
func GetTransactionsByUserID(c *gin.Context) {
	userID := c.Param("id")
	var accounts []models.Account

	// Fetch all accounts for the given user
	if err := config.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "User not found"})
		return
	}

	// Fetch transactions for the user's accounts
	var transactions []models.Transaction
	if err := config.DB.Where("account_id IN (?)", getAccountIDs(accounts)).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch transactions"})
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

// @Summary      Get transactions by account number
// @Description  Retrieves all transactions for a specific account number.
// @Tags         transactions
// @Param        account_no  path      string  true  "Account Number"
// @Produce      json
// @Success      200  {array}   models.Transaction
// @Failure      404  {object}  models.ErrorResponse  "Account not found"
// @Failure      500  {object}  models.ErrorResponse  "Failed to fetch transactions"
// @Router       /accounts/{account_no}/transactions [get]
func GetTransactionsByAccountNo(c *gin.Context) {
	accountNo := c.Param("account_no")
	var account models.Account

	// Fetch the account by account number
	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Account not found"})
		return
	}

	// Fetch transactions for the given account
	var transactions []models.Transaction
	if err := config.DB.Where("account_id = ?", account.ID).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
