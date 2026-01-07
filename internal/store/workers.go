package store

import (
	"context"

	"github.com/google/uuid"
)

func (s *Store) UpsertWorkerHeartbeat(
	ctx context.Context,
	workerID uuid.UUID,
	workerCapacity int,
) error {
	_, err := s.connectionPool.Exec(ctx, `
		INSERT INTO workers (
			id,
			last_heartbeat,
			capacity
		)
		VALUES ($1, now(), $2)
		ON CONFLICT (id)
		DO UPDATE
		SET last_heartbeat = now(),
			capacity = EXCLUDED.capacity
	`,
		workerID,
		workerCapacity,
	)

	return err
}

func (s *Store) RegisterWorker(
	ctx context.Context,
	workerID uuid.UUID,
	capacity int,
) error {
	_, err := s.connectionPool.Exec(ctx,
		`
		INSERT INTO workers (id, capacity, last_heartbeat)
		VALUES ($1, $2, now())
		ON CONFLICT (id) DO UPDATE
		SET capacity = EXCLUDED.capacity,
			last_heartbeat = now()
		`,
		workerID,
		capacity,
	)
	return err
}

func (s *Store) HeartbeatWorker(
	ctx context.Context,
	workerID uuid.UUID,
) error {
	_, err := s.connectionPool.Exec(ctx, `
	UPDATE workers
	SET last_heartbeat = now()
	WHERE id = $1
	`,
		workerID,
	)
	return err
}
