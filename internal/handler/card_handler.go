package handler

import (
	"encoding/json"
	"net/http"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"
)

type CardHandler struct {
	cardService         service.CardService
	notificationService service.NotificationService
}

func NewCardHandler(
	cardService service.CardService,
	notificationService service.NotificationService,
) *CardHandler {
	return &CardHandler{
		cardService:         cardService,
		notificationService: notificationService,
	}
}

func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	var req models.CreateCardRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	card, err := h.cardService.CreateCard(userID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, card)
}

func (h *CardHandler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	cards, err := h.cardService.GetUserCards(userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) PayByCard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	email := middleware.GetEmailFromContext(r)

	var req models.CardPaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.cardService.PayByCard(userID, req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response := models.CardPaymentResponse{
		Account:   account,
		EmailSent: false,
	}

	if h.notificationService != nil && email != "" {
		if err := h.notificationService.SendPaymentEmail(email, req.Amount); err != nil {
			response.EmailError = err.Error()
		} else {
			response.EmailSent = true
		}
	}

	writeJSON(w, http.StatusOK, response)
}
