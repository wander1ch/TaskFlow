package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/sotremont/taskflow/internal/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetTask(ctx context.Context, taskID uuid.UUID) (*dto.TaskResponse, error)
	ListTasks(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]dto.TaskResponse, error)
	UpdateTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	AssignTask(ctx context.Context, operatorID, taskID, assigneeID uuid.UUID) error
	GetAnalytics(ctx context.Context, teamID uuid.UUID) (*dto.TeamAnalyticsResponse, error)
	GetHistory(ctx context.Context, taskID uuid.UUID) ([]dto.TaskEventResponse, error)
}

type taskService struct {
	taskRepo  repository.TaskRepository
	teamRepo  repository.TeamRepository
	eventRepo repository.EventRepository
	userRepo  repository.UserRepository
}

func NewTaskService(taskRepo repository.TaskRepository, teamRepo repository.TeamRepository, eventRepo repository.EventRepository, userRepo repository.UserRepository) TaskService {
	return &taskService{taskRepo: taskRepo, teamRepo: teamRepo, eventRepo: eventRepo, userRepo: userRepo}
}

func (s *taskService) CreateTask(ctx context.Context, userID uuid.UUID, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	// Check if user is member of the team and has permission
	member, err := s.teamRepo.GetMember(ctx, req.TeamID, userID)
	if err != nil || (member.Role != domain.RoleOwner && member.Role != domain.RoleManager) {
		return nil, fmt.Errorf("unauthorized: only owners and managers can create tasks")
	}

	task := &domain.Task{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
	}
	if task.Status == "" {
		task.Status = domain.StatusTodo
	}
	if task.Priority == "" {
		task.Priority = domain.PriorityMedium
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Create event
	_ = s.eventRepo.Create(ctx, &domain.TaskEvent{
		TaskID: task.ID,
		UserID: userID,
		Action: "task_created",
	})

	return s.mapToResponse(ctx, task), nil
}

func (s *taskService) GetTask(ctx context.Context, taskID uuid.UUID) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return s.mapToResponse(ctx, task), nil
}

func (s *taskService) ListTasks(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]dto.TaskResponse, error) {
	memberships, err := s.teamRepo.GetMembersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	allowedTeams := make(map[uuid.UUID]struct{}, len(memberships))
	for _, m := range memberships {
		allowedTeams[m.TeamID] = struct{}{}
	}

	tasks, err := s.taskRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		if _, ok := allowedTeams[t.TeamID]; !ok {
			continue
		}
		resp = append(resp, *s.mapToResponse(ctx, &t))
	}
	return resp, nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// RBRB
	member, err := s.teamRepo.GetMember(ctx, task.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// Rule: Owner/Manager can update anything. Member can only update if they are assignee.
	if member.Role == domain.RoleMember && (task.AssigneeID == nil || *task.AssigneeID != userID) {
		return nil, fmt.Errorf("unauthorized: members can only update their own tasks")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		old := string(task.Status)
		task.Status = *req.Status
		_ = s.eventRepo.Create(ctx, &domain.TaskEvent{TaskID: task.ID, UserID: userID, Action: "status_changed", OldValue: old, NewValue: string(*req.Status)})
	}
	if req.Priority != nil {
		old := string(task.Priority)
		task.Priority = *req.Priority
		_ = s.eventRepo.Create(ctx, &domain.TaskEvent{TaskID: task.ID, UserID: userID, Action: "priority_changed", OldValue: old, NewValue: string(*req.Priority)})
	}
	if req.AssigneeID != nil {
		old := ""
		if task.AssigneeID != nil {
			old = task.AssigneeID.String()
		}
		task.AssigneeID = req.AssigneeID
		_ = s.eventRepo.Create(ctx, &domain.TaskEvent{TaskID: task.ID, UserID: userID, Action: "assignee_changed", OldValue: old, NewValue: req.AssigneeID.String()})
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return s.mapToResponse(ctx, task), nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	member, err := s.teamRepo.GetMember(ctx, task.TeamID, userID)
	if err != nil || (member.Role != domain.RoleOwner && member.Role != domain.RoleManager) {
		return fmt.Errorf("unauthorized")
	}

	return s.taskRepo.Delete(ctx, taskID)
}

func (s *taskService) AssignTask(ctx context.Context, operatorID, taskID, assigneeID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	member, err := s.teamRepo.GetMember(ctx, task.TeamID, operatorID)
	if err != nil || (member.Role != domain.RoleOwner && member.Role != domain.RoleManager) {
		return fmt.Errorf("unauthorized")
	}

	old := ""
	if task.AssigneeID != nil {
		old = task.AssigneeID.String()
	}
	task.AssigneeID = &assigneeID

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	_ = s.eventRepo.Create(ctx, &domain.TaskEvent{
		TaskID:   task.ID,
		UserID:   operatorID,
		Action:   "task_assigned",
		OldValue: old,
		NewValue: assigneeID.String(),
	})

	return nil
}

func (s *taskService) GetAnalytics(ctx context.Context, teamID uuid.UUID) (*dto.TeamAnalyticsResponse, error) {
	total, completed, active, rate, err := s.taskRepo.GetAnalytics(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return &dto.TeamAnalyticsResponse{
		TotalTasks:     total,
		CompletedTasks: completed,
		ActiveTasks:    active,
		CompletionRate: rate,
	}, nil
}

func (s *taskService) GetHistory(ctx context.Context, taskID uuid.UUID) ([]dto.TaskEventResponse, error) {
	events, err := s.eventRepo.ListByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.TaskEventResponse, len(events))
	for i, e := range events {
		resp[i] = dto.TaskEventResponse{
			ID:        e.ID,
			TaskID:    e.TaskID,
			UserID:    e.UserID,
			Action:    e.Action,
			OldValue:  e.OldValue,
			NewValue:  e.NewValue,
			CreatedAt: e.CreatedAt,
		}
	}
	return resp, nil
}

func (s *taskService) mapToResponse(ctx context.Context, t *domain.Task) *dto.TaskResponse {
	user, _ := s.userRepo.GetByID(ctx, t.CreatedBy)
	email := ""
	if user != nil {
		email = user.Email
	}

	team, _ := s.teamRepo.GetByID(ctx, t.TeamID)
	teamName := ""
	if team != nil {
		teamName = team.Name
	}

	return &dto.TaskResponse{
		ID:             t.ID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		AssigneeID:     t.AssigneeID,
		CreatedBy:      email,
		TeamName:       teamName,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
