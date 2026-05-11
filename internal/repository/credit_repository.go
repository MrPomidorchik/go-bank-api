package repository

import (
	"bank-api/internal/apperror"
	"database/sql"
	"errors"
	"math"

	"bank-api/internal/models"
)

type CreditRepository interface {
	CreateCredit(userID string, credit *models.Credit, schedules []models.PaymentSchedule) error
	GetUserCredits(userID string) ([]models.Credit, error)
	GetPaymentSchedule(userID, creditID string) ([]models.PaymentSchedule, error)
	ProcessDuePayments() (*models.ProcessPaymentsResult, error)
}

type creditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) CreditRepository {
	return &creditRepository{db: db}
}

func (r *creditRepository) CreateCredit(userID string, credit *models.Credit, schedules []models.PaymentSchedule) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	accountQuery := `
		SELECT id
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`

	var accountID string
	err = tx.QueryRow(accountQuery, credit.AccountID, userID).Scan(&accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("account not found")
	}
	if err != nil {
		return err
	}

	creditQuery := `
		INSERT INTO credits (
			user_id,
			account_id,
			amount,
			annual_rate,
			term_months,
			monthly_payment,
			remaining_amount,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		RETURNING id, status, created_at, updated_at
	`

	err = tx.QueryRow(
		creditQuery,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.AnnualRate,
		credit.TermMonths,
		credit.MonthlyPayment,
		credit.RemainingAmount,
	).Scan(
		&credit.ID,
		&credit.Status,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for i := range schedules {
		schedules[i].CreditID = credit.ID

		scheduleQuery := `
			INSERT INTO payment_schedules (
				credit_id,
				user_id,
				account_id,
				payment_number,
				due_date,
				amount,
				principal_amount,
				interest_amount,
				penalty_amount,
				status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 'planned')
		`

		_, err = tx.Exec(
			scheduleQuery,
			schedules[i].CreditID,
			schedules[i].UserID,
			schedules[i].AccountID,
			schedules[i].PaymentNumber,
			schedules[i].DueDate,
			schedules[i].Amount,
			schedules[i].PrincipalAmount,
			schedules[i].InterestAmount,
		)
		if err != nil {
			return err
		}
	}

	updateBalanceQuery := `
		UPDATE accounts
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`

	_, err = tx.Exec(updateBalanceQuery, credit.Amount, credit.AccountID, userID)
	if err != nil {
		return err
	}

	transactionQuery := `
		INSERT INTO transactions (user_id, account_id, type, amount)
		VALUES ($1, $2, 'credit_disbursement', $3)
	`

	_, err = tx.Exec(transactionQuery, userID, credit.AccountID, credit.Amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *creditRepository) GetUserCredits(userID string) ([]models.Credit, error) {
	query := `
		SELECT
			id,
			user_id,
			account_id,
			amount,
			annual_rate,
			term_months,
			monthly_payment,
			remaining_amount,
			status,
			created_at,
			updated_at
		FROM credits
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []models.Credit

	for rows.Next() {
		var credit models.Credit

		err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.AnnualRate,
			&credit.TermMonths,
			&credit.MonthlyPayment,
			&credit.RemainingAmount,
			&credit.Status,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		credits = append(credits, credit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credits, nil
}

func (r *creditRepository) GetPaymentSchedule(userID, creditID string) ([]models.PaymentSchedule, error) {
	query := `
		SELECT
			id,
			credit_id,
			user_id,
			account_id,
			payment_number,
			due_date,
			amount,
			principal_amount,
			interest_amount,
			penalty_amount,
			status,
			paid_at,
			created_at
		FROM payment_schedules
		WHERE user_id = $1 AND credit_id = $2
		ORDER BY payment_number ASC
	`

	rows, err := r.db.Query(query, userID, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.PaymentSchedule

	for rows.Next() {
		var schedule models.PaymentSchedule

		err := rows.Scan(
			&schedule.ID,
			&schedule.CreditID,
			&schedule.UserID,
			&schedule.AccountID,
			&schedule.PaymentNumber,
			&schedule.DueDate,
			&schedule.Amount,
			&schedule.PrincipalAmount,
			&schedule.InterestAmount,
			&schedule.PenaltyAmount,
			&schedule.Status,
			&schedule.PaidAt,
			&schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return nil, apperror.NotFound("payment schedule not found")
	}

	return schedules, nil
}

type duePaymentRow struct {
	ID              string
	CreditID        string
	UserID          string
	AccountID       string
	Amount          float64
	PrincipalAmount float64
	PenaltyAmount   float64
	Status          string
	Balance         float64
}

func (r *creditRepository) ProcessDuePayments() (*models.ProcessPaymentsResult, error) {
	query := `
		SELECT id
		FROM payment_schedules
		WHERE status IN ('planned', 'overdue')
		  AND due_date <= CURRENT_DATE
		ORDER BY due_date ASC, payment_number ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scheduleIDs []string

	for rows.Next() {
		var id string

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		scheduleIDs = append(scheduleIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &models.ProcessPaymentsResult{}

	for _, scheduleID := range scheduleIDs {
		status, err := r.processSingleDuePayment(scheduleID)
		if err != nil {
			return nil, err
		}

		result.ProcessedPayments++

		switch status {
		case "paid":
			result.PaidPayments++
		case "overdue":
			result.OverduePayments++
		}
	}

	return result, nil
}

func (r *creditRepository) processSingleDuePayment(scheduleID string) (string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	payment, err := r.getDuePaymentForUpdate(tx, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}

		return "", err
	}

	totalPayment := roundRepositoryMoney(payment.Amount + payment.PenaltyAmount)

	if payment.Balance >= totalPayment {
		if err := r.payCreditSchedule(tx, payment, totalPayment); err != nil {
			return "", err
		}

		if err := tx.Commit(); err != nil {
			return "", err
		}

		return "paid", nil
	}

	if err := r.markPaymentOverdue(tx, payment); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return "overdue", nil
}

func (r *creditRepository) getDuePaymentForUpdate(tx *sql.Tx, scheduleID string) (*duePaymentRow, error) {
	query := `
		SELECT
			ps.id,
			ps.credit_id,
			ps.user_id,
			ps.account_id,
			ps.amount,
			ps.principal_amount,
			ps.penalty_amount,
			ps.status,
			a.balance
		FROM payment_schedules ps
		JOIN accounts a ON a.id = ps.account_id
		JOIN credits c ON c.id = ps.credit_id
		WHERE ps.id = $1
		  AND ps.status IN ('planned', 'overdue')
		  AND ps.due_date <= CURRENT_DATE
		FOR UPDATE OF ps, a, c
	`

	payment := &duePaymentRow{}

	err := tx.QueryRow(query, scheduleID).Scan(
		&payment.ID,
		&payment.CreditID,
		&payment.UserID,
		&payment.AccountID,
		&payment.Amount,
		&payment.PrincipalAmount,
		&payment.PenaltyAmount,
		&payment.Status,
		&payment.Balance,
	)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *creditRepository) payCreditSchedule(tx *sql.Tx, payment *duePaymentRow, totalPayment float64) error {
	updateAccountQuery := `
		UPDATE accounts
		SET balance = balance - $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	if _, err := tx.Exec(updateAccountQuery, totalPayment, payment.AccountID); err != nil {
		return err
	}

	updateScheduleQuery := `
		UPDATE payment_schedules
		SET status = 'paid',
		    paid_at = NOW()
		WHERE id = $1
	`

	if _, err := tx.Exec(updateScheduleQuery, payment.ID); err != nil {
		return err
	}

	updateCreditQuery := `
		UPDATE credits
		SET remaining_amount = GREATEST(remaining_amount - $1, 0),
		    status = CASE
		        WHEN GREATEST(remaining_amount - $1, 0) = 0 THEN 'closed'
		        ELSE 'active'
		    END,
		    updated_at = NOW()
		WHERE id = $2
	`

	if _, err := tx.Exec(updateCreditQuery, payment.PrincipalAmount, payment.CreditID); err != nil {
		return err
	}

	transactionQuery := `
		INSERT INTO transactions (user_id, account_id, type, amount)
		VALUES ($1, $2, 'credit_payment', $3)
	`

	if _, err := tx.Exec(transactionQuery, payment.UserID, payment.AccountID, totalPayment); err != nil {
		return err
	}

	return nil
}

func (r *creditRepository) markPaymentOverdue(tx *sql.Tx, payment *duePaymentRow) error {
	penaltyToAdd := 0.0

	if payment.Status == "planned" {
		penaltyToAdd = roundRepositoryMoney(payment.Amount * 0.10)
	}

	updateScheduleQuery := `
		UPDATE payment_schedules
		SET status = 'overdue',
		    penalty_amount = penalty_amount + $1
		WHERE id = $2
	`

	if _, err := tx.Exec(updateScheduleQuery, penaltyToAdd, payment.ID); err != nil {
		return err
	}

	updateCreditQuery := `
		UPDATE credits
		SET status = 'overdue',
		    updated_at = NOW()
		WHERE id = $1
	`

	if _, err := tx.Exec(updateCreditQuery, payment.CreditID); err != nil {
		return err
	}

	return nil
}

func roundRepositoryMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
