package service

import (
	"bank-api/internal/validator"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

type AccountService interface {
	CreateAccount(userID string, req models.CreateAccountRequest) (*models.Account, error)
	GetUserAccounts(userID string) ([]models.Account, error)
	Deposit(userID, accountID string, req models.MoneyOperationRequest) (*models.Account, error)
	Withdraw(userID, accountID string, req models.MoneyOperationRequest) (*models.Account, error)
	Transfer(userID string, req models.TransferRequest) (*models.Account, error)
}

type accountService struct {
	accountRepo repository.AccountRepository
}

func NewAccountService(accountRepo repository.AccountRepository) AccountService {
	return &accountService{
		accountRepo: accountRepo,
	}
}

func (s *accountService) CreateAccount(userID string, req models.CreateAccountRequest) (*models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	currency := req.Currency
	if currency == "" {
		currency = "RUB"
	}

	if err := validator.Currency(currency); err != nil {
		return nil, err
	}

	account := &models.Account{
		UserID:        userID,
		AccountNumber: generateAccountNumber(),
		Balance:       0,
		Currency:      currency,
	}

	if err := s.accountRepo.Create(account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *accountService) GetUserAccounts(userID string) ([]models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	return s.accountRepo.GetByUserID(userID)
}

func (s *accountService) Deposit(userID, accountID string, req models.MoneyOperationRequest) (*models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(accountID, "account id"); err != nil {
		return nil, err
	}

	if err := validator.Amount(req.Amount); err != nil {
		return nil, err
	}

	return s.accountRepo.Deposit(userID, accountID, req.Amount)
}

func (s *accountService) Withdraw(userID, accountID string, req models.MoneyOperationRequest) (*models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(accountID, "account id"); err != nil {
		return nil, err
	}

	if err := validator.Amount(req.Amount); err != nil {
		return nil, err
	}

	return s.accountRepo.Withdraw(userID, accountID, req.Amount)
}

func generateAccountNumber() string {
	result := "40817810"

	for i := 0; i < 12; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			result += "0"
			continue
		}

		result += fmt.Sprintf("%d", n.Int64())
	}

	return result
}

func (s *accountService) Transfer(userID string, req models.TransferRequest) (*models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(req.FromAccountID, "from account id"); err != nil {
		return nil, err
	}

	if req.ToAccountNumber == "" {
		return nil, errors.New("to account number is required")
	}

	if err := validator.Amount(req.Amount); err != nil {
		return nil, err
	}

	return s.accountRepo.Transfer(userID, req)
}
