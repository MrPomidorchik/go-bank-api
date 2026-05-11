package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"bank-api/internal/models"
)

type TransactionRepository interface {
	GetUserTransactions(userID string, filter models.TransactionFilter) ([]models.Transaction, error)
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

func (r *transactionRepository) GetUserTransactions(
	userID string,
	filter models.TransactionFilter,
) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, account_id, type, amount, created_at
		FROM transactions
		WHERE user_id = $1
	`

	args := []any{userID}
	argIndex := 2

	var conditions []string

	if filter.AccountID != "" {
		conditions = append(conditions, fmt.Sprintf("account_id = $%d", argIndex))
		args = append(args, filter.AccountID)
		argIndex++
	}

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, filter.Type)
		argIndex++
	}

	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("created_at < ($%d::date + INTERVAL '1 day')", argIndex))
		args = append(args, filter.DateTo)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.AccountID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
