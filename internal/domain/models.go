package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Role string

const (
	RoleOwner   Role = "owner"
	RoleManager Role = "manager"
	RoleMember  Role = "member"
)

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	TeamID uuid.UUID `json:"team_id"`
	UserID uuid.UUID `json:"user_id"`
	Role   Role      `json:"role"`
}

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

type Task struct {
	ID             uuid.UUID    `json:"id"`
	TeamID         uuid.UUID    `json:"team_id"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Status         TaskStatus   `json:"status"`
	Priority       TaskPriority `json:"priority"`
	AssigneeID     *uuid.UUID   `json:"assignee_id"`
	CreatedBy      uuid.UUID    `json:"created_by"`
	CreatedByEmail string       `json:"created_by_email"`
	TeamName       string       `json:"team_name"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskEvent struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}
