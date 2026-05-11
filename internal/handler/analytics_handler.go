package handler

import (
	"net/http"
	"strconv"

	"bank-api/internal/middleware"
	"bank-api/internal/service"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) GetMonthlyAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)

	analytics, err := h.analyticsService.GetMonthlyAnalytics(userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

func (h *AnalyticsHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	accountID := mux.Vars(r)["accountId"]

	days := 30

	daysParam := r.URL.Query().Get("days")
	if daysParam != "" {
		parsedDays, err := strconv.Atoi(daysParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid days parameter")
			return
		}

		days = parsedDays
	}

	prediction, err := h.analyticsService.PredictBalance(userID, accountID, days)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, prediction)
}
