package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/Vin-Jex/job-orchestrator/internal/store"
	"github.com/google/uuid"
)

type Worker struct {
	id       uuid.UUID
	capacity int
	store    *store.Store
	logger   *slog.Logger
}

func New(id uuid.UUID, capacity int, storeLayer *store.Store, logger *slog.Logger) *Worker {
	return &Worker{
		id:       id,
		capacity: capacity,
		store:    storeLayer,
		logger:   logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.store.RegisterWorker(ctx, w.id, w.capacity); err != nil {
		return err
	}

	go w.runExecutor(ctx)

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

func (w *Worker) runExecutor(ctx context.Context) {
	semaphore := make(chan struct{}, w.capacity)

	for {
		select {
		case <-ctx.Done():
			return
		case semaphore <- struct{}{}:
			go func() {
				defer func() { <-semaphore }()

				jobID, payload, err := w.store.AcquireScheduledJobForWorker(ctx, w.id)
				if err != nil {
					time.Sleep(300 * time.Millisecond)
					return
				}
				w.logger.Info("job picked up", "job_id", jobID.String(), "worker_id", w.id.String())

				jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				cancelWatcherDone := make(chan struct{})

				go func() {
					defer close(cancelWatcherDone)

					ticker := time.NewTicker(500 * time.Millisecond)
					defer ticker.Stop()

					for {
						select {
						case <-jobCtx.Done():
							return
						case <-ticker.C:
							cancelled, err := w.store.IsJobCancelled(ctx, jobID)
							if err != nil {
								continue
							}
							if cancelled {
								cancel()
								w.logger.Info("job cancelled during execution", "job_id", jobID.String(), "worker_id", w.id.String())
								return
							}
						}
					}
				}()

				select {
				case <-jobCtx.Done():
					return
				case <-time.After(1 * time.Second):
				default:
				}

				_ = payload

				time.Sleep(1 * time.Second)

				cancelled, err := w.store.IsJobCancelled(ctx, jobID)
				if err == nil && cancelled {
					w.logger.Info("job cancelled during execution", "job_id", jobID.String(), "worker_id", w.id.String())
					return
				}

				_ = w.store.MarkJobCompleted(jobCtx, jobID)

				w.logger.Info("job completed", "job_id", jobID.String(), "worker_id", w.id.String())

			}()
		}
	}
}
