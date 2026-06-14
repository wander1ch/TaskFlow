package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotremont/taskflow/internal/domain"
)

type eventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) EventRepository {
	return &eventRepo{pool: pool}
}

func (r *eventRepo) Create(ctx context.Context, event *domain.TaskEvent) error {
	query := `INSERT INTO task_events (task_id, user_id, action, old_value, new_value) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, event.TaskID, event.UserID, event.Action, event.OldValue, event.NewValue).Scan(&event.ID, &event.CreatedAt)
}

func (r *eventRepo) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.TaskEvent, error) {
	query := `SELECT id, task_id, user_id, action, old_value, new_value, created_at FROM task_events WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.TaskEvent
	for rows.Next() {
		var e domain.TaskEvent
		if err := rows.Scan(&e.ID, &e.TaskID, &e.UserID, &e.Action, &e.OldValue, &e.NewValue, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
