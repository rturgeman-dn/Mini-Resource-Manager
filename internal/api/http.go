package api

import (
	"net/http"
	"Mini-Resource-Manager/internal/core"
)

// Handler holds all HTTP handlers and their dependencies
type Handler struct {
	store *core.Store
}

// NewHandler creates a new Handler instance with the given store
func NewHandler(store *core.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// RegisterRoutes registers all HTTP routes with the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /templates", h.HandleCreateTemplate)
	mux.HandleFunc("POST /pools", h.HandleCreatePool)
	mux.HandleFunc("POST /allocate", h.HandleAllocate)
}