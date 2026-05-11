package models

type TestEmailRequest struct {
	To string `json:"to"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
