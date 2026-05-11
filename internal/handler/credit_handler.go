package handler

import (
	"encoding/json"
	"net/http"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"

	"github.com/gorilla/mux"
)

type CreditHandler struct {
	creditService service.CreditService
}

func NewCreditHandler(creditService service.CreditService) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
	}
}

func (h *CreditHandler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	var req models.CreateCreditRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	credit, err := h.creditService.CreateCredit(userID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, credit)
}

func (h *CreditHandler) GetUserCredits(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	credits, err := h.creditService.GetUserCredits(userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, credits)
}

func (h *CreditHandler) GetPaymentSchedule(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	creditID := mux.Vars(r)["creditId"]

	schedule, err := h.creditService.GetPaymentSchedule(userID, creditID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

func (h *CreditHandler) ProcessDuePayments(w http.ResponseWriter, r *http.Request) {
	result, err := h.creditService.ProcessDuePayments()
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
