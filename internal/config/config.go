package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret   string
	JWTTTLHours int

	CBRURL         string
	BankRateMargin float64

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	HMACSecret string
	PGPSecret  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtTTLHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "24"))
	if err != nil {
		return nil, err
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, err
	}

	bankRateMargin, err := strconv.ParseFloat(getEnv("BANK_RATE_MARGIN", "5"), 64)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "bank_api"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:   getEnv("JWT_SECRET", "secret"),
		JWTTTLHours: jwtTTLHours,

		CBRURL:         getEnv("CBR_URL", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"),
		BankRateMargin: bankRateMargin,

		SMTPHost:     getEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", "noreply@example.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@example.com"),

		HMACSecret: getEnv("HMAC_SECRET", "change_me_hmac_secret"),
		PGPSecret:  getEnv("PGP_SECRET", "change_me_pgp_secret"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
