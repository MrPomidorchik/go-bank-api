package handler

import (
	"encoding/json"
	"net/http"

	"bank-api/internal/models"
	"bank-api/internal/service"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	var req models.TestEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.notificationService.SendTestEmail(req.To); err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "email sent successfully",
	})
}
