package admin

import "net/http"

type Handler struct {
	store AdminStore
}

func NewHandler(store AdminStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /admin/request-magic-link", h.requestMagicLink)
	router.HandleFunc("GET /admin/testers", h.getTesters)
	router.HandleFunc("POST /admin/testers/{id}/approve", h.approveTester)
	router.HandleFunc("POST /admin/testers/{id}/reject", h.rejectTester)
	router.HandleFunc("GET /admin/callback", h.callback)

}

func (h *Handler) requestMagicLink(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getTesters(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) approveTester(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) rejectTester(w http.ResponseWriter, r *http.Request) {

}
