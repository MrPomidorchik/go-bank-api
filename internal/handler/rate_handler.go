package handler

import (
	"net/http"

	"bank-api/internal/models"
)

type RateProvider interface {
	GetCentralBankRate() (float64, error)
}

type RateHandler struct {
	rateProvider RateProvider
}

func NewRateHandler(rateProvider RateProvider) *RateHandler {
	return &RateHandler{
		rateProvider: rateProvider,
	}
}

func (h *RateHandler) GetCBRRate(w http.ResponseWriter, r *http.Request) {
	rate, err := h.rateProvider.GetCentralBankRate()
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, models.RateResponse{
		Rate: rate,
	})
}
