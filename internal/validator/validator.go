package validator

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func Email(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return errors.New("email is required")
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return errors.New("invalid email")
	}

	return nil
}

func Username(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return errors.New("username is required")
	}

	if len(value) < 3 {
		return errors.New("username must be at least 3 characters")
	}

	if len(value) > 50 {
		return errors.New("username must not exceed 50 characters")
	}

	if !usernameRegex.MatchString(value) {
		return errors.New("username may contain only letters, numbers and underscore")
	}

	return nil
}

func Password(value string) error {
	if value == "" {
		return errors.New("password is required")
	}

	if len(value) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	if len(value) > 72 {
		return errors.New("password must not exceed 72 characters")
	}

	return nil
}

func UUID(value string, fieldName string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return errors.New(fieldName + " is required")
	}

	if _, err := uuid.Parse(value); err != nil {
		return errors.New("invalid " + fieldName)
	}

	return nil
}

func Amount(value float64) error {
	if value <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if value > 1000000000 {
		return errors.New("amount is too large")
	}

	return nil
}

func Currency(value string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	if value != "RUB" {
		return errors.New("only RUB currency is supported")
	}

	return nil
}

func CVV(value string) error {
	value = strings.TrimSpace(value)

	if len(value) != 3 {
		return errors.New("cvv must contain 3 digits")
	}

	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return errors.New("cvv must contain only digits")
		}
	}

	return nil
}

func Days(value int) error {
	if value <= 0 {
		return errors.New("days must be greater than zero")
	}

	if value > 365 {
		return errors.New("days must not exceed 365")
	}

	return nil
}

func DateYYYYMMDD(value string, fieldName string) error {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	if _, err := time.Parse("2006-01-02", value); err != nil {
		return errors.New("invalid " + fieldName + " format, use YYYY-MM-DD")
	}

	return nil
}
