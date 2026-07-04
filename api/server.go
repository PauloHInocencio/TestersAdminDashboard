package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PauloHInocencio/testers-admin-dashboard/db"
	"github.com/PauloHInocencio/testers-admin-dashboard/email"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/admin"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/session"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/tester"
	"github.com/rs/cors"
)

type Server struct {
	addr    string
	storage *db.Storage
}

func NewServer(addr string, storage *db.Storage) *Server {
	return &Server{
		addr:    ":" + addr,
		storage: storage,
	}
}

func (s *Server) Run() error {
	router := http.NewServeMux()
	v1 := http.NewServeMux()

	// Register handlers inside v1
	testersStore := tester.NewStore(s.storage.Queries)
	testersHandler := tester.NewHandler(testersStore)
	testersHandler.RegisterRoutes(v1)

	emailService := email.NewService()
	webBaseURL := os.Getenv("WEB_BASE_URL")
	sessionStore := session.NewStore(s.storage.Queries)
	adminStore := admin.NewStore(s.storage.Queries, s.storage.DB)
	adminHandler := admin.NewHandler(adminStore, testersStore, sessionStore, emailService, webBaseURL)
	adminHandler.RegisterRoutes(v1)

	// Attach v1 to main router
	router.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 0 || allowedOrigins[0] == "" {
		allowedOrigins = []string{"http://localhost:8081"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	httpServer := http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server has started %s", s.addr)

	return httpServer.ListenAndServe()
}
