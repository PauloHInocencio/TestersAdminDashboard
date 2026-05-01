package api

import (
	"log"
	"net/http"
	"time"

	"github.com/PauloHInocencio/testers-admin-dashboard/db"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/admin"
	"github.com/PauloHInocencio/testers-admin-dashboard/services/tester"
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

	adminStore := admin.NewStore(s.storage.Queries, s.storage.DB)
	adminHandler := admin.NewHandler(adminStore)
	adminHandler.RegisterRoutes(v1)

	// Attach v1 to main router
	router.Handle("/api/v1", http.StripPrefix("/api/v1", v1))

	httpServer := http.Server{
		Addr:              s.addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server has started %s", s.addr)

	return httpServer.ListenAndServe()
}
