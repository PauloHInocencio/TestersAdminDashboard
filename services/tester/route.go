package tester

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/PauloHInocencio/testers-admin-dashboard/middleware"
	"github.com/PauloHInocencio/testers-admin-dashboard/models"
	"github.com/PauloHInocencio/testers-admin-dashboard/utils"
)

type Handler struct {
	store TestersStore
}

func NewHandler(store TestersStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	logging := middleware.GetLoggingMiddleware()
	router.HandleFunc("POST /testers/signup", logging(h.signupTester))
}

func (h *Handler) signupTester(w http.ResponseWriter, r *http.Request) {
	var req models.TesterSignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Invalid email address",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Name is required",
		})
	}

	validPlatforms := map[string]bool{"android": true, "ios": true, "both": true}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if !validPlatforms[platform] {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Invalid platform. Must be: android, ios, or both",
		})
		return
	}

	if err := h.store.CreateSignup(r.Context(), email, name, platform); err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to create signup",
		})
	}

	utils.JSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Thanks! You've been added to the tester waitlist.",
	})
}
