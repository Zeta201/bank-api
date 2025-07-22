package handlers

import (
	"bank-app/config"
	"bank-app/models"
	"bank-app/rabbitmq"
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

	c.JSON(http.StatusCreated, models.AccountCreatedResponse{
		Message:     "Account created successfully",
		AccountID:   account.ID,
		AccountNo:   account.AccountNo,
		Balance:     account.Balance,
		AccountType: account.AccountType,
	})
}

// @Summary      Deposit money into an account
// @Description  Deposits a specified amount into the account identified by its account number
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        account_no  path      string                     true  "Account number"
// @Param        request     body      models.TransactionRequest  true  "Deposit amount"
// @Success      200         {object}  models.TransactionResponse
// @Failure      400         {object}  models.ErrorResponse
// @Failure      404         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/{account_no}/deposit [post]
func Deposit(c *gin.Context) {
	accountNo := c.Param("account_no")
	userID := c.MustGet("userID").(uint)
	var request models.TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid amount"})
		return
	}

	// Retrieve the user to get phone number (or email)
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch user info"})
		return
	}

	// Find the account
	var account models.Account
	if err := config.DB.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Account not found"})
		return
	}

	// Add the deposit amount to the balance
	account.Balance += request.Amount
	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update balance"})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to log transaction"})
		return
	}

	_ = rabbitmq.Publish(map[string]interface{}{
		"type":      "deposit",
		"status":    "success",
		"user_id":   account.UserID,
		"amount":    request.Amount,
		"accountNo": account.AccountNo,
		"timestamp": time.Now().UTC(),
		"to_email":  user.Email,
	})

	c.JSON(http.StatusOK, models.TransactionResponse{
		Message: "Deposit successful",
		Balance: account.Balance,
	})
}

// @Summary      Withdraw money from an account
// @Description  Withdraws a specified amount from the authenticated user's account
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        account_no  path      string                     true  "Account number"
// @Param        request     body      models.TransactionRequest  true  "Withdrawal amount"
// @Success      200         {object}  models.TransactionResponse
// @Failure      400         {object}  models.ErrorResponse
// @Failure      403         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/{account_no}/withdraw [post]
func Withdraw(c *gin.Context) {
	accountNo := c.Param("account_no")
	var request models.TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid or missing amount"})
		return
	}

	// Get authenticated user's ID from context
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch user info"})
		return
	}

	// Find the account AND ensure it belongs to the authenticated user
	var account models.Account
	if err := config.DB.Where("account_no = ? AND user_id = ?", accountNo, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Account not found or access denied"})
		return
	}

	// Check if the account has enough balance
	if account.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Insufficient balance"})
		return
	}

	// Deduct the amount
	account.Balance -= request.Amount
	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update balance"})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to log transaction"})
		return
	}

	_ = rabbitmq.Publish(map[string]interface{}{
		"type":      "withdraw",
		"status":    "success",
		"user_id":   account.UserID,
		"amount":    request.Amount,
		"accountNo": account.AccountNo,
		"timestamp": time.Now().UTC(),
		"to_email":  user.Email,
	})

	c.JSON(http.StatusOK, models.TransactionResponse{
		Message: "Withdrawal successful",
		Balance: account.Balance,
	})
}

func Transfer(c *gin.Context) {
	fromAccountNo := c.Param("from_account")
	toAccountNo := c.Param("to_account")

	var request models.TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid or missing amount"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var sender models.User
	if err := config.DB.First(&sender, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch user info"})
		return
	}

	// Ensure the 'from' account belongs to the logged-in user
	var fromAccount models.Account
	if err := config.DB.Where("account_no = ? AND user_id = ?", fromAccountNo, userID).First(&fromAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "You do not have access to this account"})
		return
	}

	// Lookup receiver's account
	var toAccount models.Account
	if err := config.DB.Where("account_no = ?", toAccountNo).First(&toAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Receiver account not found"})
		return
	}

	// Prevent transferring to the same account
	if fromAccount.AccountNo == toAccount.AccountNo {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Cannot transfer to the same account"})
		return
	}

	// Check sufficient balance
	if fromAccount.Balance < request.Amount {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Insufficient balance"})
		return
	}

	// Execute transfer
	fromAccount.Balance -= request.Amount
	toAccount.Balance += request.Amount

	// Use transaction to ensure atomicity
	tx := config.DB.Begin()
	if err := tx.Save(&fromAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update sender account"})
		return
	}
	if err := tx.Save(&toAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update receiver account"})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to log sender transaction"})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to log receiver transaction"})
		return
	}

	// Commit transaction
	tx.Commit()

	var receiver models.User
	if err := config.DB.First(&receiver, toAccount.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch receiver info"})
		return
	}
	_ = rabbitmq.Publish(map[string]interface{}{
		"type":      "transfer_sent",
		"status":    "success",
		"user_id":   fromAccount.UserID,
		"amount":    request.Amount,
		"from":      fromAccount.AccountNo,
		"to":        toAccount.AccountNo,
		"timestamp": time.Now().UTC(),
		"to_email":  sender.Email,
	})

	_ = rabbitmq.Publish(map[string]interface{}{
		"type":      "transfer_received",
		"status":    "success",
		"user_id":   toAccount.UserID,
		"amount":    request.Amount,
		"from":      fromAccount.AccountNo,
		"to":        toAccount.AccountNo,
		"timestamp": time.Now().UTC(),
		"to_email":  receiver.Email,
	})

	c.JSON(http.StatusOK, models.TransactionResponse{
		Message: "Transfer successful",
		Balance: fromAccount.Balance,
	})
}

// @Summary      Get all accounts for authenticated user
// @Description  Retrieves a list of all bank accounts belonging to the authenticated user
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.AccountsResponse
// @Failure      500  {object}  models.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts [get]
func GetAllAccounts(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var accounts []models.Account
	if err := config.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, models.AccountsResponse{
		Accounts: accounts,
	})
}
