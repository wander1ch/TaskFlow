package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
)

// Auth
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// User
type UserResponse struct {
	ID        uuid.UUID       `json:"id"`
	Email     string          `json:"email"`
	CreatedAt time.Time       `json:"created_at"`
	Roles     []UserTeamRole `json:"roles"`
}

type UserTeamRole struct {
	TeamID   uuid.UUID    `json:"team_id"`
	TeamName string       `json:"team_name"`
	Role     domain.Role `json:"role"`
}

type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// Team
type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
}

type TeamResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AddMemberRequest struct {
	Email string      `json:"email" binding:"required,email"`
	Role  domain.Role `json:"role" binding:"required,oneof=owner manager member"`
}

type MemberResponse struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
}

// Task
type CreateTaskRequest struct {
	TeamID      uuid.UUID           `json:"team_id" binding:"required"`
	Title       string              `json:"title" binding:"required"`
	Description string              `json:"description"`
	Status      domain.TaskStatus   `json:"status" binding:"omitempty,oneof=todo in_progress review done"`
	Priority    domain.TaskPriority `json:"priority" binding:"omitempty,oneof=low medium high critical"`
	AssigneeID  *uuid.UUID          `json:"assignee_id"`
}

type UpdateTaskRequest struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Status      *domain.TaskStatus   `json:"status" binding:"omitempty,oneof=todo in_progress review done"`
	Priority    *domain.TaskPriority `json:"priority" binding:"omitempty,oneof=low medium high critical"`
	AssigneeID  *uuid.UUID           `json:"assignee_id"`
}

type AssignTaskRequest struct {
	AssigneeID uuid.UUID `json:"assignee_id" binding:"required"`
}

type TaskResponse struct {
	ID             uuid.UUID           `json:"id"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Status         domain.TaskStatus   `json:"status"`
	Priority       domain.TaskPriority `json:"priority"`
	AssigneeID     *uuid.UUID          `json:"assignee_id"`
	CreatedBy      string              `json:"created_by"`
	TeamName       string              `json:"team_name"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// Comment
type CreateCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type CommentResponse struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// Analytics
type TeamAnalyticsResponse struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	ActiveTasks    int     `json:"active_tasks"`
	CompletionRate float64 `json:"completion_rate"`
}

// History
type TaskEventResponse struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}
