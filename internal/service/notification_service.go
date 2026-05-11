package service

import (
	"errors"
	"strings"
)

type EmailSender interface {
	SendEmail(to string, subject string, body string) error
	SendPaymentEmail(to string, amount float64) error
}

type NotificationService interface {
	SendTestEmail(to string) error
	SendPaymentEmail(to string, amount float64) error
}

type notificationService struct {
	emailSender EmailSender
}

func NewNotificationService(emailSender EmailSender) NotificationService {
	return &notificationService{
		emailSender: emailSender,
	}
}

func (s *notificationService) SendTestEmail(to string) error {
	to = strings.TrimSpace(to)

	if to == "" {
		return errors.New("email is required")
	}

	if !strings.Contains(to, "@") {
		return errors.New("invalid email")
	}

	body := `
		<h1>SMTP работает</h1>
		<p>Это тестовое письмо от REST API банковского сервиса.</p>
	`

	return s.emailSender.SendEmail(to, "Тестовое письмо Bank API", body)
}

func (s *notificationService) SendPaymentEmail(to string, amount float64) error {
	to = strings.TrimSpace(to)

	if to == "" {
		return errors.New("email is required")
	}

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	return s.emailSender.SendPaymentEmail(to, amount)
}
