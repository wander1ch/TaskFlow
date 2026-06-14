package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/sotremont/taskflow/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	GenerateToken(user *domain.User) (string, error)
}

type authService struct {
	repo      repository.UserRepository
	teamRepo  repository.TeamRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, teamRepo repository.TeamRepository, jwtSecret string) AuthService {
	return &authService{repo: repo, teamRepo: teamRepo, jwtSecret: jwtSecret}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailExists
	}
	if err != repository.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		// This part is technically redundant if err != ErrNotFound is handled, 
		// but safe to keep or just rely on the error check.
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// Fetch roles (though they should be empty for a new user)
	roles := []dto.UserTeamRole{}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			Roles:     roles,
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// Fetch user roles
	roles := []dto.UserTeamRole{}
	memberships, err := s.teamRepo.GetMembersByUserID(ctx, user.ID)
	if err == nil {
		for _, m := range memberships {
			team, err := s.teamRepo.GetByID(ctx, m.TeamID)
			if err == nil && team != nil {
				roles = append(roles, dto.UserTeamRole{
					TeamID:   m.TeamID,
					TeamName: team.Name,
					Role:     m.Role,
				})
			}
		}
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			Roles:     roles,
		},
	}, nil
}

func (s *authService) GenerateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
