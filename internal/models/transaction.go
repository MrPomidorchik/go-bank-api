package models

type TransactionFilter struct {
	AccountID string
	Type      string
	DateFrom  string
	DateTo    string
}
