package worker

import (
	"context"
	"time"

	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/google/uuid"
)

type Worker struct {
	id       uuid.UUID
	capacity int
	store    *store.Store
}

func New(id uuid.UUID, capacity int, storeLayer *store.Store) *Worker {
	return &Worker{
		id:       id,
		capacity: capacity,
		store:    storeLayer,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.store.RegisterWorker(ctx, w.id, w.capacity); err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := w.store.HeartbeatWorker(ctx, w.id); err != nil {
				return err
			}
		}
	}
}
