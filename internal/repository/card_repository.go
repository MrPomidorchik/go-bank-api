package repository

import (
	"bank-api/internal/apperror"
	"database/sql"
	"errors"

	"bank-api/internal/models"
)

type CardRepository interface {
	CreateCard(
		card *models.CardResponse,
		cvvHash string,
		cardNumberHMAC string,
		expiryHMAC string,
		pgpSecret string,
	) error

	GetUserCards(userID string, pgpSecret string) ([]models.CardResponse, error)
	AccountBelongsToUser(userID, accountID string) (bool, error)

	GetCardPaymentData(userID, cardID, pgpSecret string) (*models.CardPaymentData, error)
	PayByCard(userID, accountID string, amount float64) (*models.Account, error)
}

type cardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) AccountBelongsToUser(userID, accountID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM accounts
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, accountID, userID).Scan(&exists)

	return exists, err
}

func (r *cardRepository) CreateCard(
	card *models.CardResponse,
	cvvHash string,
	cardNumberHMAC string,
	expiryHMAC string,
	pgpSecret string,
) error {
	query := `
		INSERT INTO cards (
			user_id,
			account_id,
			card_number_encrypted,
			expiry_encrypted,
			cvv_hash,
			card_number_hmac,
			expiry_hmac
		)
		VALUES (
			$1,
			$2,
			pgp_sym_encrypt($3, $4),
			pgp_sym_encrypt($5, $4),
			$6,
			$7,
			$8
		)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		card.UserID,
		card.AccountID,
		card.CardNumber,
		pgpSecret,
		card.ExpiryDate,
		cvvHash,
		cardNumberHMAC,
		expiryHMAC,
	).Scan(
		&card.ID,
		&card.CreatedAt,
	)

	return err
}

func (r *cardRepository) GetUserCards(userID string, pgpSecret string) ([]models.CardResponse, error) {
	query := `
		SELECT
			id,
			user_id,
			account_id,
			pgp_sym_decrypt(card_number_encrypted, $2) AS card_number,
			pgp_sym_decrypt(expiry_encrypted, $2) AS expiry_date,
			created_at
		FROM cards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID, pgpSecret)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.CardResponse

	for rows.Next() {
		var card models.CardResponse

		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.AccountID,
			&card.CardNumber,
			&card.ExpiryDate,
			&card.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		card.CVV = ""

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func ErrCardAccountNotFound() error {
	return errors.New("account not found")
}

func (r *cardRepository) GetCardPaymentData(userID, cardID, pgpSecret string) (*models.CardPaymentData, error) {
	query := `
		SELECT
			id,
			user_id,
			account_id,
			pgp_sym_decrypt(card_number_encrypted, $3) AS card_number,
			pgp_sym_decrypt(expiry_encrypted, $3) AS expiry_date,
			cvv_hash,
			card_number_hmac,
			expiry_hmac
		FROM cards
		WHERE id = $1 AND user_id = $2
	`

	card := &models.CardPaymentData{}

	err := r.db.QueryRow(query, cardID, userID, pgpSecret).Scan(
		&card.ID,
		&card.UserID,
		&card.AccountID,
		&card.CardNumber,
		&card.ExpiryDate,
		&card.CVVHash,
		&card.CardNumberHMAC,
		&card.ExpiryHMAC,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("card not found")
	}

	if err != nil {
		return nil, err
	}

	return card, nil
}

func (r *cardRepository) PayByCard(userID, accountID string, amount float64) (*models.Account, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, user_id, account_number, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`

	account := &models.Account{}

	err = tx.QueryRow(query, accountID, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("account not found")
	}

	if err != nil {
		return nil, err
	}

	if account.Balance < amount {
		return nil, apperror.BadRequest("insufficient funds")
	}

	account.Balance -= amount

	updateQuery := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at
	`

	err = tx.QueryRow(updateQuery, account.Balance, account.ID).Scan(&account.UpdatedAt)
	if err != nil {
		return nil, err
	}

	transactionQuery := `
		INSERT INTO transactions (user_id, account_id, type, amount)
		VALUES ($1, $2, 'card_payment', $3)
	`

	_, err = tx.Exec(transactionQuery, userID, accountID, amount)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return account, nil
}
