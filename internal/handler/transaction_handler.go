package handler

import (
	"net/http"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	filter := models.TransactionFilter{
		AccountID: r.URL.Query().Get("account_id"),
		Type:      r.URL.Query().Get("type"),
		DateFrom:  r.URL.Query().Get("date_from"),
		DateTo:    r.URL.Query().Get("date_to"),
	}

	transactions, err := h.transactionService.GetUserTransactions(userID, filter)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transactions)
}
