package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotremont/taskflow/internal/domain"
)

var ErrNotFound = errors.New("resource not found")

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, created_at FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, created_at FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET email = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, user.Email, user.ID)
	return err
}
