package repository

import (
	"bank-api/internal/apperror"
	"database/sql"
	"errors"

	"bank-api/internal/models"
)

type AccountRepository interface {
	Create(account *models.Account) error
	GetByUserID(userID string) ([]models.Account, error)
	Deposit(userID, accountID string, amount float64) (*models.Account, error)
	Withdraw(userID, accountID string, amount float64) (*models.Account, error)
	Transfer(userID string, req models.TransferRequest) (*models.Account, error)
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, account_number, balance, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		account.UserID,
		account.AccountNumber,
		account.Balance,
		account.Currency,
	).Scan(
		&account.ID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	return err
}

func (r *accountRepository) GetByUserID(userID string) ([]models.Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account

	for rows.Next() {
		var account models.Account

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.AccountNumber,
			&account.Balance,
			&account.Currency,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *accountRepository) getAccountByNumberForUpdate(tx *sql.Tx, accountNumber string) (*models.Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, created_at, updated_at
		FROM accounts
		WHERE account_number = $1
		FOR UPDATE
	`

	account := &models.Account{}

	err := tx.QueryRow(query, accountNumber).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("destination account not found")
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *accountRepository) Deposit(userID, accountID string, amount float64) (*models.Account, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	account, err := r.getAccountForUpdate(tx, userID, accountID)
	if err != nil {
		return nil, err
	}

	account.Balance += amount

	if err := r.updateBalance(tx, account); err != nil {
		return nil, err
	}

	if err := r.createTransaction(tx, userID, accountID, "deposit", amount); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return account, nil
}

func (r *accountRepository) Withdraw(userID, accountID string, amount float64) (*models.Account, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	account, err := r.getAccountForUpdate(tx, userID, accountID)
	if err != nil {
		return nil, err
	}

	if account.Balance < amount {
		return nil, apperror.BadRequest("insufficient funds")
	}

	account.Balance -= amount

	if err := r.updateBalance(tx, account); err != nil {
		return nil, err
	}

	if err := r.createTransaction(tx, userID, accountID, "withdraw", amount); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return account, nil
}

func (r *accountRepository) Transfer(userID string, req models.TransferRequest) (*models.Account, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	sourceAccount, err := r.getAccountForUpdate(tx, userID, req.FromAccountID)
	if err != nil {
		return nil, err
	}

	if sourceAccount.AccountNumber == req.ToAccountNumber {
		return nil, apperror.BadRequest("cannot transfer to the same account")
	}

	destinationAccount, err := r.getAccountByNumberForUpdate(tx, req.ToAccountNumber)
	if err != nil {
		return nil, err
	}

	if sourceAccount.Balance < req.Amount {
		return nil, apperror.BadRequest("insufficient funds")
	}

	sourceAccount.Balance -= req.Amount
	destinationAccount.Balance += req.Amount

	if err := r.updateBalance(tx, sourceAccount); err != nil {
		return nil, err
	}

	if err := r.updateBalance(tx, destinationAccount); err != nil {
		return nil, err
	}

	if err := r.createTransaction(tx, userID, sourceAccount.ID, "transfer_out", req.Amount); err != nil {
		return nil, err
	}

	if err := r.createTransaction(tx, destinationAccount.UserID, destinationAccount.ID, "transfer_in", req.Amount); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return sourceAccount, nil
}

func (r *accountRepository) getAccountForUpdate(tx *sql.Tx, userID, accountID string) (*models.Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`

	account := &models.Account{}

	err := tx.QueryRow(query, accountID, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("account not found")
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *accountRepository) updateBalance(tx *sql.Tx, account *models.Account) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at
	`

	return tx.QueryRow(query, account.Balance, account.ID).Scan(&account.UpdatedAt)
}

func (r *accountRepository) createTransaction(tx *sql.Tx, userID, accountID, operationType string, amount float64) error {
	query := `
		INSERT INTO transactions (user_id, account_id, type, amount)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tx.Exec(query, userID, accountID, operationType, amount)
	return err
}
