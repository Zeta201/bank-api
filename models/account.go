package models

import "github.com/jinzhu/gorm"

type Account struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	AccountNo   string  `json:"account_no" gorm:"unique;not null"`
	Balance     float64 `json:"balance"`
	AccountType string  `json:"account_type"`
}
