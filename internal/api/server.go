package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Vin-Jex/job-orchestrator/internal/observability"
	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/google/uuid"
)

type Server struct {
	store  *store.Store
	mux    *http.ServeMux
	logger *slog.Logger
}

type loggerKey struct{}

func NewServer(storeLayer *store.Store, logger *slog.Logger) *Server {
	server := &Server{
		store:  storeLayer,
		mux:    http.NewServeMux(),
		logger: logger,
	}

	server.registerRoutes()

	return server
}

func (s *Server) Handler() http.Handler {
	return s.withRequestID(s.mux)
}

func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), observability.RequestIDKey(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) withRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger

		if requestID, ok := observability.RequestIDFromContext(r.Context()); ok {
			logger = logger.With("request_id", requestID)
		}

		ctx := context.WithValue(r.Context(), loggerKey{}, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}
