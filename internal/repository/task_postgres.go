package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotremont/taskflow/internal/domain"
)

type taskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &taskRepo{pool: pool}
}

func (r *taskRepo) Create(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO tasks (team_id, title, description, status, priority, assignee_id, created_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, task.TeamID, task.Title, task.Description, task.Status, task.Priority, task.AssigneeID, task.CreatedBy).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *taskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t := &domain.Task{}
	query := `SELECT id, team_id, title, description, status, priority, assignee_id, created_by, created_at, updated_at FROM tasks WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *taskRepo) List(ctx context.Context, filters map[string]interface{}) ([]domain.Task, error) {
	query := `SELECT id, team_id, title, description, status, priority, assignee_id, created_by, created_at, updated_at FROM tasks WHERE 1=1`
	var args []interface{}
	i := 1

	if teamID, ok := filters["team_id"]; ok {
		query += fmt.Sprintf(" AND team_id = $%d", i)
		args = append(args, teamID)
		i++
	}
	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, status)
		i++
	}
	if priority, ok := filters["priority"]; ok {
		query += fmt.Sprintf(" AND priority = $%d", i)
		args = append(args, priority)
		i++
	}
	if assigneeID, ok := filters["assignee_id"]; ok {
		query += fmt.Sprintf(" AND assignee_id = $%d", i)
		args = append(args, assigneeID)
		i++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *taskRepo) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4, assignee_id = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6 RETURNING updated_at`
	return r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.Priority, task.AssigneeID, task.ID).Scan(&task.UpdatedAt)
}

func (r *taskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *taskRepo) GetAnalytics(ctx context.Context, teamID uuid.UUID) (total, completed, active int, rate float64, err error) {
	query := `SELECT 
				COUNT(*) as total,
				COUNT(*) FILTER (WHERE status = 'done') as completed,
				COUNT(*) FILTER (WHERE status != 'done') as active
			  FROM tasks WHERE team_id = $1`
	err = r.pool.QueryRow(ctx, query, teamID).Scan(&total, &completed, &active)
	if err != nil {
		return
	}
	if total > 0 {
		rate = float64(completed) / float64(total) * 100
	}
	return
}
