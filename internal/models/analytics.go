package models

type MonthlyAnalytics struct {
	Income               float64 `json:"income"`
	Expenses             float64 `json:"expenses"`
	Net                  float64 `json:"net"`
	CreditMonthlyPayment float64 `json:"credit_monthly_payment"`
	CreditLoadPercent    float64 `json:"credit_load_percent"`
	Month                string  `json:"month"`
}

type BalancePrediction struct {
	AccountID        string  `json:"account_id"`
	CurrentBalance   float64 `json:"current_balance"`
	PlannedPayments  float64 `json:"planned_payments"`
	PredictedBalance float64 `json:"predicted_balance"`
	Days             int     `json:"days"`
}
