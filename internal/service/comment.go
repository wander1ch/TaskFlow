package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/sotremont/taskflow/internal/repository"
)

type CommentService interface {
	CreateComment(ctx context.Context, userID, taskID uuid.UUID, text string) (*dto.CommentResponse, error)
	ListComments(ctx context.Context, taskID uuid.UUID) ([]dto.CommentResponse, error)
	DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
	teamRepo    repository.TeamRepository
}

func NewCommentService(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository, teamRepo repository.TeamRepository) CommentService {
	return &commentService{commentRepo: commentRepo, taskRepo: taskRepo, teamRepo: teamRepo}
}

func (s *commentService) CreateComment(ctx context.Context, userID, taskID uuid.UUID, text string) (*dto.CommentResponse, error) {
	// Check access to task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	_, err = s.teamRepo.GetMember(ctx, task.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: must be team member to comment")
	}

	comment := &domain.Comment{
		TaskID: taskID,
		UserID: userID,
		Text:   text,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return &dto.CommentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		UserID:    comment.UserID,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (s *commentService) ListComments(ctx context.Context, taskID uuid.UUID) ([]dto.CommentResponse, error) {
	comments, err := s.commentRepo.ListByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.CommentResponse, len(comments))
	for i, c := range comments {
		resp[i] = dto.CommentResponse{
			ID:        c.ID,
			TaskID:    c.TaskID,
			UserID:    c.UserID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt,
		}
	}
	return resp, nil
}

func (s *commentService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment.UserID != userID {
		return fmt.Errorf("unauthorized: only comment author can delete it")
	}

	return s.commentRepo.Delete(ctx, commentID)
}
