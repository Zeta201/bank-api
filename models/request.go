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

type LoginResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token"`
	User    UserResponse `json:"user"`
}

type AccountCreatedResponse struct {
	Message     string  `json:"message"`
	AccountID   uint    `json:"account_id"`
	AccountNo   string  `json:"account_no"`
	Balance     float64 `json:"balance"`
	AccountType string  `json:"account_type"`
}

type TransactionResponse struct {
	Message string  `json:"message"`
	Balance float64 `json:"balance"`
}

type TransactionRequest struct {
	Amount float64 `json:"amount"`
}

type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}
