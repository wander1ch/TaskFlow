package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotremont/taskflow/internal/domain"
)

type teamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &teamRepo{pool: pool}
}

func (r *teamRepo) Create(ctx context.Context, team *domain.Team) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO teams (name, owner_id) VALUES ($1, $2) RETURNING id, created_at`
	err = tx.QueryRow(ctx, query, team.Name, team.OwnerID).Scan(&team.ID, &team.CreatedAt)
	if err != nil {
		return err
	}

	queryMember := `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, queryMember, team.ID, team.OwnerID, domain.RoleOwner)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *teamRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	team := &domain.Team{}
	query := `SELECT id, name, owner_id, created_at FROM teams WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return team, err
}

func (r *teamRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Team, error) {
	query := `SELECT t.id, t.name, t.owner_id, t.created_at FROM teams t 
			  JOIN team_members tm ON t.id = tm.team_id WHERE tm.user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *teamRepo) List(ctx context.Context) ([]domain.Team, error) {
	query := `SELECT id, name, owner_id, created_at FROM teams`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *teamRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *teamRepo) AddMember(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error {
	query := `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, teamID, userID, role)
	return err
}

func (r *teamRepo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, teamID, userID)
	return err
}

func (r *teamRepo) GetMembers(ctx context.Context, teamID uuid.UUID) ([]domain.TeamMember, error) {
	query := `SELECT team_id, user_id, role FROM team_members WHERE team_id = $1`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *teamRepo) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*domain.TeamMember, error) {
	var m domain.TeamMember
	query := `SELECT team_id, user_id, role FROM team_members WHERE team_id = $1 AND user_id = $2`
	err := r.pool.QueryRow(ctx, query, teamID, userID).Scan(&m.TeamID, &m.UserID, &m.Role)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &m, err
}

func (r *teamRepo) GetMembersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.TeamMember, error) {
	query := `SELECT team_id, user_id, role FROM team_members WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *teamRepo) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error {
	query := `UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3`
	_, err := r.pool.Exec(ctx, query, role, teamID, userID)
	return err
}
