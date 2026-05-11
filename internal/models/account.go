package models

import "time"

type Account struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

type MoneyOperationRequest struct {
	Amount float64 `json:"amount"`
}

type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	AccountID string    `json:"account_id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type TransferRequest struct {
	FromAccountID   string  `json:"from_account_id"`
	ToAccountNumber string  `json:"to_account_number"`
	Amount          float64 `json:"amount"`
}
