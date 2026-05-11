package models

import "time"

type CreateCreditRequest struct {
	AccountID  string  `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
	AnnualRate float64 `json:"annual_rate,omitempty"`
}

type Credit struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	AccountID       string    `json:"account_id"`
	Amount          float64   `json:"amount"`
	AnnualRate      float64   `json:"annual_rate"`
	TermMonths      int       `json:"term_months"`
	MonthlyPayment  float64   `json:"monthly_payment"`
	RemainingAmount float64   `json:"remaining_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PaymentSchedule struct {
	ID              string     `json:"id"`
	CreditID        string     `json:"credit_id"`
	UserID          string     `json:"user_id"`
	AccountID       string     `json:"account_id"`
	PaymentNumber   int        `json:"payment_number"`
	DueDate         time.Time  `json:"due_date"`
	Amount          float64    `json:"amount"`
	PrincipalAmount float64    `json:"principal_amount"`
	InterestAmount  float64    `json:"interest_amount"`
	PenaltyAmount   float64    `json:"penalty_amount"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ProcessPaymentsResult struct {
	ProcessedPayments int `json:"processed_payments"`
	PaidPayments      int `json:"paid_payments"`
	OverduePayments   int `json:"overdue_payments"`
}
