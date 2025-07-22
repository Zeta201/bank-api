package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model `swaggerignore:"true"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email" gorm:"unique;not null"`
	Password   string    `json:"password"`
	Phone      string    `json:"user_phone"`
	Accounts   []Account `json:"accounts"`
}
