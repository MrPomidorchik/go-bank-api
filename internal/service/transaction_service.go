package service

import (
	"bank-api/internal/models"
	"bank-api/internal/repository"
	"bank-api/internal/validator"
	"errors"
)

type TransactionService interface {
	GetUserTransactions(userID string, filter models.TransactionFilter) ([]models.Transaction, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

func (s *transactionService) GetUserTransactions(
	userID string,
	filter models.TransactionFilter,
) ([]models.Transaction, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if filter.AccountID != "" {
		if err := validator.UUID(filter.AccountID, "account id"); err != nil {
			return nil, err
		}
	}

	if filter.Type != "" && !isAllowedTransactionType(filter.Type) {
		return nil, errors.New("invalid transaction type")
	}

	if err := validator.DateYYYYMMDD(filter.DateFrom, "date_from"); err != nil {
		return nil, err
	}

	if err := validator.DateYYYYMMDD(filter.DateTo, "date_to"); err != nil {
		return nil, err
	}

	return s.transactionRepo.GetUserTransactions(userID, filter)
}

func isAllowedTransactionType(transactionType string) bool {
	allowedTypes := map[string]bool{
		"deposit":             true,
		"withdraw":            true,
		"transfer_in":         true,
		"transfer_out":        true,
		"card_payment":        true,
		"credit_disbursement": true,
		"credit_payment":      true,
	}

	return allowedTypes[transactionType]
}
