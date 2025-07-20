package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Transaction struct {
	gorm.Model      `swaggerignore:"true"`
	TransactionType string    `json:"transaction_type"` // Deposit, Withdrawal, Transfer
	Amount          float64   `json:"amount"`
	AccountID       uint      `json:"account_id"`                // Account that initiated the transaction
	FromAccountID   *uint     `json:"from_account_id,omitempty"` // For transfers, the originating account
	ToAccountID     *uint     `json:"to_account_id,omitempty"`   // For transfers, the receiving account
	Status          string    `json:"status"`                    // Success or failure
	TransactionDate time.Time `json:"transaction_date"`
}
