// Scheduler is a long-running control-plane component responsible for
// leasing eligible jobs and coordinating their execution.
//
// Responsibilities:
//   - Acquire leases for PENDING jobs
//   - Detect and recover expired leases
//   - Ensure at-most-once job execution
//   - Remain stateless and crash-safe
//
// The scheduler does not execute jobs.
// It only coordinates state transitions via the store layer.
//
// This binary is intended to be run as a standalone process.
// Multiple schedulers may run concurrently.
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
	"github.com/vin-jex/job-orchestrator/internal/scheduler"
	"github.com/vin-jex/job-orchestrator/internal/store"
)

func main() {
	_ = godotenv.Load()

	logger := observability.NewLogger("scheduler")

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

	schedulerID := uuid.New()
	s := scheduler.New(schedulerID, storeLayer, logger)

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

	go s.Run(ctx)
	<-ctx.Done()
}
