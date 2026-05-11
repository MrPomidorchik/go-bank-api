package service

import (
	"bank-api/internal/validator"
	"errors"
	"math"
	"time"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

type CreditService interface {
	CreateCredit(userID string, req models.CreateCreditRequest) (*models.Credit, error)
	GetUserCredits(userID string) ([]models.Credit, error)
	GetPaymentSchedule(userID, creditID string) ([]models.PaymentSchedule, error)
	ProcessDuePayments() (*models.ProcessPaymentsResult, error)
}

type RateProvider interface {
	GetCentralBankRate() (float64, error)
}

type creditService struct {
	creditRepo   repository.CreditRepository
	rateProvider RateProvider
}

func NewCreditService(
	creditRepo repository.CreditRepository,
	rateProvider RateProvider,
) CreditService {
	return &creditService{
		creditRepo:   creditRepo,
		rateProvider: rateProvider,
	}
}

func (s *creditService) CreateCredit(userID string, req models.CreateCreditRequest) (*models.Credit, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(req.AccountID, "account id"); err != nil {
		return nil, err
	}

	if err := validator.Amount(req.Amount); err != nil {
		return nil, err
	}

	if req.TermMonths <= 0 {
		return nil, errors.New("term months must be greater than zero")
	}

	if req.TermMonths > 360 {
		return nil, errors.New("term months must not exceed 360")
	}

	annualRate := req.AnnualRate
	if annualRate <= 0 {
		if s.rateProvider == nil {
			return nil, errors.New("rate provider is not configured")
		}

		rate, err := s.rateProvider.GetCentralBankRate()
		if err != nil {
			return nil, err
		}

		annualRate = rate
	}

	monthlyPayment := calculateAnnuityPayment(req.Amount, annualRate, req.TermMonths)

	credit := &models.Credit{
		UserID:          userID,
		AccountID:       req.AccountID,
		Amount:          roundMoney(req.Amount),
		AnnualRate:      annualRate,
		TermMonths:      req.TermMonths,
		MonthlyPayment:  monthlyPayment,
		RemainingAmount: roundMoney(req.Amount),
	}

	schedules := buildPaymentSchedule(credit)

	if err := s.creditRepo.CreateCredit(userID, credit, schedules); err != nil {
		return nil, err
	}

	return credit, nil
}

func (s *creditService) GetUserCredits(userID string) ([]models.Credit, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	return s.creditRepo.GetUserCredits(userID)
}

func (s *creditService) GetPaymentSchedule(userID, creditID string) ([]models.PaymentSchedule, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(creditID, "credit id"); err != nil {
		return nil, err
	}

	return s.creditRepo.GetPaymentSchedule(userID, creditID)
}

func calculateAnnuityPayment(amount float64, annualRate float64, termMonths int) float64 {
	monthlyRate := annualRate / 100 / 12

	if monthlyRate == 0 {
		return roundMoney(amount / float64(termMonths))
	}

	pow := math.Pow(1+monthlyRate, float64(termMonths))

	payment := amount * monthlyRate * pow / (pow - 1)

	return roundMoney(payment)
}

func buildPaymentSchedule(credit *models.Credit) []models.PaymentSchedule {
	monthlyRate := credit.AnnualRate / 100 / 12
	remaining := credit.Amount

	schedules := make([]models.PaymentSchedule, 0, credit.TermMonths)

	for i := 1; i <= credit.TermMonths; i++ {
		interest := roundMoney(remaining * monthlyRate)
		principal := roundMoney(credit.MonthlyPayment - interest)

		if i == credit.TermMonths {
			principal = roundMoney(remaining)
		}

		paymentAmount := roundMoney(principal + interest)
		remaining = roundMoney(remaining - principal)

		schedule := models.PaymentSchedule{
			UserID:          credit.UserID,
			AccountID:       credit.AccountID,
			PaymentNumber:   i,
			DueDate:         time.Now().AddDate(0, i, 0),
			Amount:          paymentAmount,
			PrincipalAmount: principal,
			InterestAmount:  interest,
			PenaltyAmount:   0,
			Status:          "planned",
		}

		schedules = append(schedules, schedule)
	}

	return schedules
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func (s *creditService) ProcessDuePayments() (*models.ProcessPaymentsResult, error) {
	return s.creditRepo.ProcessDuePayments()
}
