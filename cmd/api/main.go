package main

import (
	smtpintegration "bank-api/internal/integration/smtp"

	"fmt"
	"log"
	"net/http"
	"time"

	"bank-api/internal/config"
	"bank-api/internal/handler"
	"bank-api/internal/integration/cbr"
	"bank-api/internal/middleware"
	"bank-api/internal/repository"
	"bank-api/internal/scheduler"
	"bank-api/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTTTLHours)
	authHandler := handler.NewAuthHandler(authService)

	accountRepo := repository.NewAccountRepository(db)
	accountService := service.NewAccountService(accountRepo)
	accountHandler := handler.NewAccountHandler(accountService)

	smtpClient := smtpintegration.NewClient(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	notificationService := service.NewNotificationService(smtpClient)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	cardRepo := repository.NewCardRepository(db)
	cardService := service.NewCardService(cardRepo, cfg.HMACSecret, cfg.PGPSecret)
	cardHandler := handler.NewCardHandler(cardService, notificationService)

	cbrClient := cbr.NewClient(cfg.CBRURL, cfg.BankRateMargin)
	rateHandler := handler.NewRateHandler(cbrClient)

	creditRepo := repository.NewCreditRepository(db)
	creditService := service.NewCreditService(creditRepo, cbrClient)
	creditHandler := handler.NewCreditHandler(creditService)

	analyticsRepo := repository.NewAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	transactionRepo := repository.NewTransactionRepository(db)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	creditScheduler := scheduler.NewCreditScheduler(creditService, 12*time.Hour)
	creditScheduler.Start()

	r := mux.NewRouter()

	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.SecurityHeadersMiddleware)

	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	protected.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"` + userID + `"}`))
	}).Methods(http.MethodGet)

	protected.HandleFunc("/accounts", accountHandler.CreateAccount).Methods(http.MethodPost)
	protected.HandleFunc("/accounts", accountHandler.GetUserAccounts).Methods(http.MethodGet)
	protected.HandleFunc("/accounts/{accountId}/deposit", accountHandler.Deposit).Methods(http.MethodPost)
	protected.HandleFunc("/accounts/{accountId}/withdraw", accountHandler.Withdraw).Methods(http.MethodPost)
	protected.HandleFunc("/transfer", accountHandler.Transfer).Methods(http.MethodPost)

	protected.HandleFunc("/cards", cardHandler.CreateCard).Methods(http.MethodPost)
	protected.HandleFunc("/cards", cardHandler.GetUserCards).Methods(http.MethodGet)
	protected.HandleFunc("/cards/pay", cardHandler.PayByCard).Methods(http.MethodPost)

	protected.HandleFunc("/credits", creditHandler.CreateCredit).Methods(http.MethodPost)
	protected.HandleFunc("/credits", creditHandler.GetUserCredits).Methods(http.MethodGet)
	protected.HandleFunc("/credits/process-payments", creditHandler.ProcessDuePayments).Methods(http.MethodPost)
	protected.HandleFunc("/credits/{creditId}/schedule", creditHandler.GetPaymentSchedule).Methods(http.MethodGet)
	protected.HandleFunc("/rates/cbr", rateHandler.GetCBRRate).Methods(http.MethodGet)

	protected.HandleFunc("/analytics", analyticsHandler.GetMonthlyAnalytics).Methods(http.MethodGet)
	protected.HandleFunc("/accounts/{accountId}/predict", analyticsHandler.PredictBalance).Methods(http.MethodGet)

	protected.HandleFunc("/transactions", transactionHandler.GetUserTransactions).Methods(http.MethodGet)

	protected.HandleFunc("/notifications/test-email", notificationHandler.SendTestEmail).Methods(http.MethodPost)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("Bank API started on port %s", cfg.ServerPort)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
