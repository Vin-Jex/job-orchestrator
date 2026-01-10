package api

import (
	"log/slog"
	"net/http"

	"github.com/Vin-Jex/job-orchestrator/internal/store"
)

type Server struct {
	store  *store.Store
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewServer(storeLayer *store.Store, logger *slog.Logger) *Server {
	server := &Server{
		store: storeLayer,
		mux:   http.NewServeMux(),
		logger: logger,
	}

	server.registerRoutes()

	return server
}

func (s *Server) Handler() http.Handler {
	return s.mux
}
