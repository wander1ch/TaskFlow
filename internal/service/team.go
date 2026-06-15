package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/cache"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/sotremont/taskflow/internal/repository"
)

type TeamService interface {
	CreateTeam(ctx context.Context, userID uuid.UUID, req dto.CreateTeamRequest) (*dto.TeamResponse, error)
	GetTeam(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error)
	ListTeams(ctx context.Context, userID uuid.UUID) ([]dto.TeamResponse, error)
	ListAllTeams(ctx context.Context) ([]dto.TeamResponse, error)
	AddMember(ctx context.Context, operatorID, teamID uuid.UUID, req dto.AddMemberRequest) error
	JoinTeam(ctx context.Context, userID, teamID uuid.UUID) error
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]dto.MemberResponse, error)
	DeleteTeam(ctx context.Context, operatorID, teamID uuid.UUID) error
	CheckRole(ctx context.Context, teamID, userID uuid.UUID, requiredRoles ...domain.Role) (bool, error)
	UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uuid.UUID, newRole domain.Role) error
}

func (s *teamService) ListAllTeams(ctx context.Context) ([]dto.TeamResponse, error) {
	teams, err := s.teamRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.TeamResponse, len(teams))
	for i, t := range teams {
		resp[i] = dto.TeamResponse{
			ID:        t.ID,
			Name:      t.Name,
			OwnerID:   t.OwnerID,
			CreatedAt: t.CreatedAt,
		}
	}
	return resp, nil
}

func (s *teamService) JoinTeam(ctx context.Context, userID, teamID uuid.UUID) error {
	return s.teamRepo.AddMember(ctx, teamID, userID, domain.RoleMember)
}

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	cache    cache.Cache
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository, cache cache.Cache) TeamService {
	return &teamService{teamRepo: teamRepo, userRepo: userRepo, cache: cache}
}

func (s *teamService) CreateTeam(ctx context.Context, userID uuid.UUID, req dto.CreateTeamRequest) (*dto.TeamResponse, error) {
	team := &domain.Team{
		Name:    req.Name,
		OwnerID: userID,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	return &dto.TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		OwnerID:   team.OwnerID,
		CreatedAt: team.CreatedAt,
	}, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error) {
	cacheKey := fmt.Sprintf("team:%s", teamID)
	var cached dto.TeamResponse
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil && cached.ID != uuid.Nil {
		return &cached, nil
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	resp := &dto.TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		OwnerID:   team.OwnerID,
		CreatedAt: team.CreatedAt,
	}

	_ = s.cache.Set(ctx, cacheKey, resp, 15*time.Minute)
	return resp, nil
}

func (s *teamService) ListTeams(ctx context.Context, userID uuid.UUID) ([]dto.TeamResponse, error) {
	teams, err := s.teamRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.TeamResponse, len(teams))
	for i, t := range teams {
		resp[i] = dto.TeamResponse{
			ID:        t.ID,
			Name:      t.Name,
			OwnerID:   t.OwnerID,
			CreatedAt: t.CreatedAt,
		}
	}
	return resp, nil
}

func (s *teamService) AddMember(ctx context.Context, operatorID, teamID uuid.UUID, req dto.AddMemberRequest) error {
	// RBAC: Only owner can add members
	isOwner, err := s.CheckRole(ctx, teamID, operatorID, domain.RoleOwner)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized: only team owner can add members")
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	return s.teamRepo.AddMember(ctx, teamID, user.ID, req.Role)
}

func (s *teamService) GetMembers(ctx context.Context, teamID uuid.UUID) ([]dto.MemberResponse, error) {
	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.MemberResponse, len(members))
	for i, m := range members {
		user, _ := s.userRepo.GetByID(ctx, m.UserID)
		email := ""
		if user != nil {
			email = user.Email
		}
		resp[i] = dto.MemberResponse{
			UserID: m.UserID,
			Email:  email,
			Role:   m.Role,
		}
	}
	return resp, nil
}

func (s *teamService) DeleteTeam(ctx context.Context, operatorID, teamID uuid.UUID) error {
	isOwner, err := s.CheckRole(ctx, teamID, operatorID, domain.RoleOwner)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized: only team owner can delete team")
	}

	if err := s.teamRepo.Delete(ctx, teamID); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("team:%s", teamID))
	return nil
}

func (s *teamService) CheckRole(ctx context.Context, teamID, userID uuid.UUID, requiredRoles ...domain.Role) (bool, error) {
	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return false, err
	}

	for _, m := range members {
		if m.UserID == userID {
			for _, requiredRole := range requiredRoles {
				if m.Role == requiredRole {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (s *teamService) UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uuid.UUID, newRole domain.Role) error {
	isOwner, err := s.CheckRole(ctx, teamID, operatorID, domain.RoleOwner)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized: only team owner can change roles")
	}

	return s.teamRepo.UpdateMemberRole(ctx, teamID, targetUserID, newRole)
}
