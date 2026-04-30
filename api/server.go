package api

import (
	"log"
	"net/http"
	"time"

	"github.com/PauloHInocencio/testers-admin-dashboard/db"
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

	// TODO: setup handlers

	httpServer := http.Server{
		Addr:              s.addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server has started %s", s.addr)

	return httpServer.ListenAndServe()
}
