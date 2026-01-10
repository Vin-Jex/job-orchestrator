package observability

import (
	"log/slog"
	"os"
)

func NewLogger(component string) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
	).With("component", component)
}
