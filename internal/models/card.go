package models

import "time"

type CardResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	AccountID  string    `json:"account_id"`
	CardNumber string    `json:"card_number"`
	ExpiryDate string    `json:"expiry_date"`
	CVV        string    `json:"cvv,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateCardRequest struct {
	AccountID string `json:"account_id"`
}

type CardPaymentRequest struct {
	CardID string  `json:"card_id"`
	CVV    string  `json:"cvv"`
	Amount float64 `json:"amount"`
}

type CardPaymentData struct {
	ID             string
	UserID         string
	AccountID      string
	CardNumber     string
	ExpiryDate     string
	CVVHash        string
	CardNumberHMAC string
	ExpiryHMAC     string
}

type CardPaymentResponse struct {
	Account    *Account `json:"account"`
	EmailSent  bool     `json:"email_sent"`
	EmailError string   `json:"email_error,omitempty"`
}
