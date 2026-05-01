package tester

import "net/http"

type Handler struct {
	store TestersStore
}

func NewHandler(store TestersStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /testers/signup", h.signupTester)
}

func (h *Handler) signupTester(w http.ResponseWriter, r *http.Request) {

}
