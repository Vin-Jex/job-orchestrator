package scheduler

import (
	"log/slog"

	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/google/uuid"
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
