package service

import (
	"bank-api/internal/validator"
	"errors"
	"time"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

type AnalyticsService interface {
	GetMonthlyAnalytics(userID string) (*models.MonthlyAnalytics, error)
	PredictBalance(userID, accountID string, days int) (*models.BalancePrediction, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

func (s *analyticsService) GetMonthlyAnalytics(userID string) (*models.MonthlyAnalytics, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	now := time.Now()

	startDate := time.Date(
		now.Year(),
		now.Month(),
		1,
		0,
		0,
		0,
		0,
		now.Location(),
	)

	endDate := startDate.AddDate(0, 1, 0)

	return s.analyticsRepo.GetMonthlyAnalytics(userID, startDate, endDate)
}

func (s *analyticsService) PredictBalance(
	userID string,
	accountID string,
	days int,
) (*models.BalancePrediction, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if err := validator.UUID(accountID, "account id"); err != nil {
		return nil, err
	}

	if days <= 0 {
		days = 30
	}

	if err := validator.Days(days); err != nil {
		return nil, err
	}

	return s.analyticsRepo.PredictBalance(userID, accountID, days)
}
