package models

type SignUpRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

type AccountRequest struct {
	AccountType    string  `json:"account_type"`
	InitialBalance float64 `json:"initial_balance"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type AmountRequest struct {
	Amount float64 `json:"amount"`
}
