package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
	
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]domain.TeamMember, error)
	GetMember(ctx context.Context, teamID, userID uuid.UUID) (*domain.TeamMember, error)
	GetMembersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	List(ctx context.Context, filters map[string]interface{}) ([]domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	GetAnalytics(ctx context.Context, teamID uuid.UUID) (total, completed, active int, rate float64, err error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
}

type EventRepository interface {
	Create(ctx context.Context, event *domain.TaskEvent) error
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.TaskEvent, error)
}
