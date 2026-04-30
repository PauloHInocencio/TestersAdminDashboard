package models

import "time"

type TesterSignup struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Platform   string     `json:"platform"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt,omitempty"`
	ApprovedAt *time.Time `json:"approvedAt,omitempty"`
	RejectedAt *time.Time `json:"rejectedAt,omitempty"`
	InvitedAt  *time.Time `json:"invitedAt,omitempty"`
}

type TesterSignupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

type TesterSignupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RequestMagicLinkRequest struct {
	Email string `json:"email"`
}
