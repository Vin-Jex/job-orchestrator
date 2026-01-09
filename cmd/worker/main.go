package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/Vin-Jex/job-orchestrator/internal/worker"
)

func main() {
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

	w := worker.New(workerID, 4, storeLayer)

	if err := w.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
