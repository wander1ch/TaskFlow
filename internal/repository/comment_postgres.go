package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotremont/taskflow/internal/domain"
)

type commentRepo struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) CommentRepository {
	return &commentRepo{pool: pool}
}

func (r *commentRepo) Create(ctx context.Context, comment *domain.Comment) error {
	query := `INSERT INTO comments (task_id, user_id, text) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, comment.TaskID, comment.UserID, comment.Text).Scan(&comment.ID, &comment.CreatedAt)
}

func (r *commentRepo) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error) {
	query := `SELECT id, task_id, user_id, text, created_at FROM comments WHERE task_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *commentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *commentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	c := &domain.Comment{}
	query := `SELECT id, task_id, user_id, text, created_at FROM comments WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Text, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return c, err
}
