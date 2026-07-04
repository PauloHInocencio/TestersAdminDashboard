package models

import "time"

type TesterSignup struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Platform   string     `json:"platform"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	ApprovedAt *time.Time `json:"approvedAt"`
	RejectedAt *time.Time `json:"rejectedAt"`
	InvitedAt  *time.Time `json:"invitedAt"`
}

type TesterSignupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ListOfTestersResponse struct {
	Data []TesterSignup `json:"data"`
}

type RequestMagicLinkRequest struct {
	Email string `json:"email"`
}
