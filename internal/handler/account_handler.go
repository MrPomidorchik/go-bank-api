package handler

import (
	"encoding/json"
	"net/http"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"

	"github.com/gorilla/mux"
)

type AccountHandler struct {
	accountService service.AccountService
}

func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	var req models.CreateAccountRequest

	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	account, err := h.accountService.CreateAccount(userID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

func (h *AccountHandler) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	accounts, err := h.accountService.GetUserAccounts(userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	accountID := mux.Vars(r)["accountId"]

	var req models.MoneyOperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.accountService.Deposit(userID, accountID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	accountID := mux.Vars(r)["accountId"]

	var req models.MoneyOperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.accountService.Withdraw(userID, accountID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	var req models.TransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.accountService.Transfer(userID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, account)
}
