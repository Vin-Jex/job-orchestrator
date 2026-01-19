// Worker is a long-running executor responsible for running leased jobs.
//
// Responsibilities:
//   - Poll for scheduled jobs assigned to this worker
//   - Execute jobs concurrently using a bounded worker pool
//   - Respect cancellation and timeouts via context
//   - Report completion or failure back to the control plane
//
// Workers do not make scheduling decisions.
// They only act on authoritative state from the store.
//
// This binary is intended to be run as a standalone process.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/vin-jex/job-orchestrator/internal/observability"
	"github.com/vin-jex/job-orchestrator/internal/store"
	"github.com/vin-jex/job-orchestrator/internal/worker"
)

func main() {
	_ = godotenv.Load()

	logger := observability.NewLogger("worker")
	workerID := uuid.New()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	storeLayer, err := store.NewStore(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer storeLayer.Close()

	w := worker.New(workerID, 4, storeLayer, logger)

	// Render keepalive HTTP server (infrastructure hack)
	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		)); err != nil {
			log.Fatal(err)
		}
	}()

	if err := w.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
