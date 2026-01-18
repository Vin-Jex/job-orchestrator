package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/Vin-Jex/job-orchestrator/internal/observability"
	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/Vin-Jex/job-orchestrator/internal/worker"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	if err := w.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
