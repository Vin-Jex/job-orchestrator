package scheduler

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/vin-jex/job-orchestrator/internal/store"
)

type Scheduler struct {
	id     uuid.UUID
	store  *store.Store
	logger *slog.Logger
}

func New(
	id uuid.UUID,
	storeLayer *store.Store,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		id:     id,
		store:  storeLayer,
		logger: logger,
	}
}
