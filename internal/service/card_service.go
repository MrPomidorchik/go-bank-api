package service

import (
	"bank-api/internal/apperror"
	"bank-api/internal/validator"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	cryptoutil "bank-api/internal/crypto"
	"bank-api/internal/models"
	"bank-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type CardService interface {
	CreateCard(userID string, req models.CreateCardRequest) (*models.CardResponse, error)
	GetUserCards(userID string) ([]models.CardResponse, error)
	PayByCard(userID string, req models.CardPaymentRequest) (*models.Account, error)
}

type cardService struct {
	cardRepo   repository.CardRepository
	hmacSecret string
	pgpSecret  string
}

func NewCardService(
	cardRepo repository.CardRepository,
	hmacSecret string,
	pgpSecret string,
) CardService {
	return &cardService{
		cardRepo:   cardRepo,
		hmacSecret: hmacSecret,
		pgpSecret:  pgpSecret,
	}
}

func (s *cardService) CreateCard(userID string, req models.CreateCardRequest) (*models.CardResponse, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(req.AccountID, "account id"); err != nil {
		return nil, err
	}

	exists, err := s.cardRepo.AccountBelongsToUser(userID, req.AccountID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, apperror.NotFound("account not found")
	}

	cardNumber := generateCardNumber()
	expiryDate := generateExpiryDate()
	cvv := generateCVV()

	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	cardNumberHMAC := cryptoutil.ComputeHMAC(cardNumber, []byte(s.hmacSecret))
	expiryHMAC := cryptoutil.ComputeHMAC(expiryDate, []byte(s.hmacSecret))

	card := &models.CardResponse{
		UserID:     userID,
		AccountID:  req.AccountID,
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
	}

	err = s.cardRepo.CreateCard(
		card,
		string(cvvHash),
		cardNumberHMAC,
		expiryHMAC,
		s.pgpSecret,
	)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) GetUserCards(userID string) ([]models.CardResponse, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	return s.cardRepo.GetUserCards(userID, s.pgpSecret)
}

func generateCardNumber() string {
	prefix := "2202"

	number := prefix

	for len(number) < 15 {
		number += randomDigit()
	}

	checkDigit := calculateLuhnCheckDigit(number)

	return number + fmt.Sprintf("%d", checkDigit)
}

func calculateLuhnCheckDigit(number string) int {
	sum := 0
	double := true

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return (10 - (sum % 10)) % 10
}

func generateExpiryDate() string {
	return time.Now().AddDate(3, 0, 0).Format("01/06")
}

func generateCVV() string {
	result := ""

	for i := 0; i < 3; i++ {
		result += randomDigit()
	}

	return result
}

func randomDigit() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		return "0"
	}

	return fmt.Sprintf("%d", n.Int64())
}

func (s *cardService) PayByCard(userID string, req models.CardPaymentRequest) (*models.Account, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(req.CardID, "card id"); err != nil {
		return nil, err
	}

	if err := validator.CVV(req.CVV); err != nil {
		return nil, err
	}

	if err := validator.Amount(req.Amount); err != nil {
		return nil, err
	}

	card, err := s.cardRepo.GetCardPaymentData(userID, req.CardID, s.pgpSecret)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(card.CVVHash), []byte(req.CVV))
	if err != nil {
		return nil, apperror.BadRequest("invalid cvv")
	}

	cardNumberHMAC := cryptoutil.ComputeHMAC(card.CardNumber, []byte(s.hmacSecret))
	if cardNumberHMAC != card.CardNumberHMAC {
		return nil, apperror.BadRequest("card number integrity check failed")
	}

	expiryHMAC := cryptoutil.ComputeHMAC(card.ExpiryDate, []byte(s.hmacSecret))
	if expiryHMAC != card.ExpiryHMAC {
		return nil, apperror.BadRequest("expiry date integrity check failed")
	}

	return s.cardRepo.PayByCard(userID, card.AccountID, req.Amount)
}
