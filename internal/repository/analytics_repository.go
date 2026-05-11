package repository

import (
	"database/sql"
	"errors"
	"time"

	"bank-api/internal/models"
)

type AnalyticsRepository interface {
	GetMonthlyAnalytics(userID string, startDate, endDate time.Time) (*models.MonthlyAnalytics, error)
	PredictBalance(userID, accountID string, days int) (*models.BalancePrediction, error)
}

type analyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetMonthlyAnalytics(
	userID string,
	startDate time.Time,
	endDate time.Time,
) (*models.MonthlyAnalytics, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE
				WHEN type IN ('deposit', 'transfer_in', 'credit_disbursement')
				THEN amount ELSE 0 END), 0) AS income,

			COALESCE(SUM(CASE
				WHEN type IN ('withdraw', 'transfer_out', 'card_payment', 'credit_payment')
				THEN amount ELSE 0 END), 0) AS expenses
		FROM transactions
		WHERE user_id = $1
		  AND created_at >= $2
		  AND created_at < $3
	`

	var income float64
	var expenses float64

	err := r.db.QueryRow(query, userID, startDate, endDate).Scan(&income, &expenses)
	if err != nil {
		return nil, err
	}

	creditQuery := `
		SELECT COALESCE(SUM(monthly_payment), 0)
		FROM credits
		WHERE user_id = $1
		  AND status IN ('active', 'overdue')
	`

	var creditMonthlyPayment float64

	err = r.db.QueryRow(creditQuery, userID).Scan(&creditMonthlyPayment)
	if err != nil {
		return nil, err
	}

	creditLoadPercent := 0.0
	if income > 0 {
		creditLoadPercent = creditMonthlyPayment / income * 100
	}

	return &models.MonthlyAnalytics{
		Income:               income,
		Expenses:             expenses,
		Net:                  income - expenses,
		CreditMonthlyPayment: creditMonthlyPayment,
		CreditLoadPercent:    creditLoadPercent,
		Month:                startDate.Format("2006-01"),
	}, nil
}

func (r *analyticsRepository) PredictBalance(
	userID string,
	accountID string,
	days int,
) (*models.BalancePrediction, error) {
	accountQuery := `
		SELECT balance
		FROM accounts
		WHERE id = $1 AND user_id = $2
	`

	var currentBalance float64

	err := r.db.QueryRow(accountQuery, accountID, userID).Scan(&currentBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("account not found")
	}
	if err != nil {
		return nil, err
	}

	paymentsQuery := `
		SELECT COALESCE(SUM(amount + penalty_amount), 0)
		FROM payment_schedules
		WHERE user_id = $1
		  AND account_id = $2
		  AND status IN ('planned', 'overdue')
		  AND due_date <= CURRENT_DATE + ($3::int)
	`

	var plannedPayments float64

	err = r.db.QueryRow(paymentsQuery, userID, accountID, days).Scan(&plannedPayments)
	if err != nil {
		return nil, err
	}

	return &models.BalancePrediction{
		AccountID:        accountID,
		CurrentBalance:   currentBalance,
		PlannedPayments:  plannedPayments,
		PredictedBalance: currentBalance - plannedPayments,
		Days:             days,
	}, nil
}
