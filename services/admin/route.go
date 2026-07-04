package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PauloHInocencio/testers-admin-dashboard/email"
	"github.com/PauloHInocencio/testers-admin-dashboard/middleware"
	"github.com/PauloHInocencio/testers-admin-dashboard/models"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/session"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/tester"
	"github.com/PauloHInocencio/testers-admin-dashboard/utils"
	"github.com/google/uuid"
)

type Handler struct {
	adminStore   AdminStore
	testerStore  tester.TestersStore
	sessionStore session.SessionStore
	emailService email.EmailService
	webBaseURL   string
}

func NewHandler(
	adminStore AdminStore,
	testerStore tester.TestersStore,
	sessionStore session.SessionStore,
	emailService email.EmailService,
	webBaseURL string,
) *Handler {
	return &Handler{
		adminStore:   adminStore,
		testerStore:  testerStore,
		sessionStore: sessionStore,
		emailService: emailService,
		webBaseURL:   webBaseURL,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	requireAdmin := middleware.RequireAdmin(h.sessionStore)
	logging := middleware.GetLoggingMiddleware()
	router.HandleFunc("POST /admin/request-magic-link", logging(h.requestMagicLink))
	router.HandleFunc("GET /admin/callback", logging(h.callback))
	router.HandleFunc("GET /admin/testers", logging(requireAdmin(h.getTesters)))
	router.HandleFunc("POST /admin/testers/{id}/approve", logging(requireAdmin(h.approveTester)))
	router.HandleFunc("POST /admin/testers/{id}/reject", logging(requireAdmin(h.rejectTester)))
	router.HandleFunc("DELETE /admin/testers/{id}/delete", logging(requireAdmin(h.deleteTester)))
}

func (h *Handler) requestMagicLink(w http.ResponseWriter, r *http.Request) {
	// Get email in the request body
	var req models.RequestMagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}
	_email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email is in the whitelist
	allowed, err := h.adminStore.IsAdminWhitelisted(r.Context(), _email)
	if err != nil || !allowed {
		// Always return success to not reveal whitelist status
		utils.JSON(w, http.StatusOK, models.ApiResponse{
			Success: true,
			Message: "If this email is authorized, a magic link will be sent.",
		})
		return
	}

	// Generate token
	token, err := utils.GenerateToken()
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}
	tokenHash := utils.HashToken(token)

	// Create magic link record
	err = h.adminStore.CreateMagicLink(r.Context(), AdminAuthParams{
		id:        uuid.New(),
		email:     _email,
		tokenHash: tokenHash,
		expiresAt: time.Now().Add(2 * time.Minute),
	})
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to create magic link",
		})
		return
	}

	// Build magic link URL
	link := fmt.Sprintf("%s/api/v1/admin/callback?token=%s", h.webBaseURL, token)

	// Send email
	if err = h.emailService.SendMagicLink(r.Context(), _email, link); err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to send email",
		})

		log.Printf("Failed to send email %e", err)
		return
	}

	utils.JSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "If this email is authorized, a magic link will be sent.",
	})

}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	// Get token from the query string
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}
	tokenHash := utils.HashToken(token)

	// Consume magic link
	_email, err := h.adminStore.ConsumeMagicLink(r.Context(), tokenHash)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Create admin session
	sessionToken, err := utils.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}
	sessionHash := utils.HashToken(sessionToken)
	expiresAt := time.Now().Add(15 * time.Minute)
	err = h.adminStore.CreateAdminSession(r.Context(), AdminAuthParams{
		id:        uuid.New(),
		email:     _email,
		tokenHash: sessionHash,
		expiresAt: expiresAt,
	})
	if err != nil {
		http.Error(w, "Failed to create sesion", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8081"
	}
	http.Redirect(w, r, frontendURL+"/admin/testers-list", http.StatusSeeOther)

}

func (h *Handler) getTesters(w http.ResponseWriter, r *http.Request) {
	testers, err := h.testerStore.ListAll(r.Context())
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to list testers",
		})
	}
	utils.JSON(w, http.StatusOK, models.ListOfTestersResponse{
		Data: testers,
	})
}

func (h *Handler) approveTester(w http.ResponseWriter, r *http.Request) {
	// Extract and validate tester id fom path
	testerModel, testerID, err := h.getTester(w, r)
	if err != nil {
		return
	}

	// Update status to approved
	if err = h.testerStore.Approve(r.Context(), testerID); err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Fail to approve tester",
		})
	}

	// Determine which email to send based on platform
	var emailErr error
	androidLink := os.Getenv("ANDROID_INVITE_LINK")
	iosLink := os.Getenv("IOS_INVITE_LINK")
	switch testerModel.Platform {
	case "android":
		emailErr = h.emailService.SendAndroidInvite(r.Context(), testerModel.Email, androidLink)
	case "ios":
		emailErr = h.emailService.SendIOSInvite(r.Context(), testerModel.Email, iosLink)
	case "both":
		androidEmailErr := h.emailService.SendAndroidInvite(r.Context(), testerModel.Email, androidLink)
		iosEmailErr := h.emailService.SendIOSInvite(r.Context(), testerModel.Email, iosLink)
		if androidEmailErr != nil || iosEmailErr != nil {
			emailErr = fmt.Errorf("failed to send invites: android=%v, ios=%v", androidEmailErr, iosEmailErr)
		}
	}

	// If email fails, log but don't fail the request (approval already succeeded)
	if emailErr != nil {
		log.Printf("Failed to send email to %s: %v", testerModel.Email, emailErr)
		_ = h.testerStore.MarkAsInvited(r.Context(), testerID)
		utils.JSON(w, http.StatusOK, models.ApiResponse{
			Success: true,
			Message: "Tester approved but email failed. Please resend invite manually",
		})
		return
	}

	// Mark as invited after successful email
	if err = h.testerStore.MarkAsInvited(r.Context(), testerID); err != nil {
		log.Printf("Failed to mark tester %s as invited: %v", testerID, err)
	}

	utils.JSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Tester approved and invite sent",
	})
}

func (h *Handler) rejectTester(w http.ResponseWriter, r *http.Request) {
	_, testerID, err := h.getTester(w, r)
	if err != nil {
		return
	}

	if err = h.testerStore.Reject(r.Context(), testerID); err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Fail to reject tester",
		})
		return
	}

	utils.JSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: fmt.Sprintf("Tester %s rejected", testerID),
	})
}

func (h *Handler) deleteTester(w http.ResponseWriter, r *http.Request) {
	_, testerID, err := h.getTester(w, r)
	if err != nil {
		return
	}

	err = h.testerStore.Delete(r.Context(), testerID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Message: "Failed to delete tester",
		})
		return
	}

	utils.JSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Tester deleted successfully",
	})
}

func (h *Handler) getTester(w http.ResponseWriter, r *http.Request) (models.TesterSignup, uuid.UUID, error) {
	// Extract and validate tester id fom path
	idStr := r.PathValue("id")
	testerID, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Message: "Invalid tester ID format",
		})
		return models.TesterSignup{}, uuid.UUID{}, nil
	}

	// Verify tester exists
	testerModel, err := h.testerStore.FindByID(r.Context(), testerID)
	if err != nil {
		utils.JSON(w, http.StatusNotFound, models.ApiResponse{
			Success: false,
			Message: "Tester not found",
		})
		return models.TesterSignup{}, uuid.UUID{}, nil
	}

	return testerModel, testerID, nil
}
