package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	user.ID = uuid.New()
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockTeamRepo struct {
	mock.Mock
}

func (m *MockTeamRepo) Create(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Team, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Team), args.Error(1)
}

func (m *MockTeamRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTeamRepo) AddMember(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error {
	args := m.Called(ctx, teamID, userID, role)
	return args.Error(0)
}

func (m *MockTeamRepo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	args := m.Called(ctx, teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepo) GetMembers(ctx context.Context, teamID uuid.UUID) ([]domain.TeamMember, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]domain.TeamMember), args.Error(1)
}

func (m *MockTeamRepo) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*domain.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TeamMember), args.Error(1)
}

func (m *MockTeamRepo) List(ctx context.Context) ([]domain.Team, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Team), args.Error(1)
}

func (m *MockTeamRepo) GetMembersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.TeamMember, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.TeamMember), args.Error(1)
}

func (m *MockTeamRepo) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role domain.Role) error {
	args := m.Called(ctx, teamID, userID, role)
	return args.Error(0)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := new(MockUserRepo)
	teamRepo := new(MockTeamRepo)
	svc := NewAuthService(repo, teamRepo, "secret")

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	existingUser := &domain.User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	repo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	resp, err := svc.Register(context.Background(), req)

	assert.ErrorIs(t, err, ErrEmailExists)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(MockUserRepo)
	teamRepo := new(MockTeamRepo)
	svc := NewAuthService(repo, teamRepo, "secret")

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
	}

	repo.On("GetByEmail", mock.Anything, email).Return(user, nil)
	teamRepo.On("GetMembersByUserID", mock.Anything, user.ID).Return([]domain.TeamMember{}, nil)

	req := dto.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	repo.AssertExpectations(t)
}
