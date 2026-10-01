package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/utils/email"
	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.Init("test")
	os.Exit(m.Run())
}

type stubUserRepo struct {
	interfaces.UserRepository
	users map[string]*models.User
}

func (s *stubUserRepo) FindByEmail(ctx context.Context, em string) (*models.User, error) {
	return s.users[em], nil
}

func (s *stubUserRepo) Create(ctx context.Context, u *models.User) error {
	u.ID = uuid.New()
	s.users[u.Email] = u
	return nil
}

func (s *stubUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	return nil
}

type stubRoleRepo struct {
	interfaces.RoleRepository
}

func (s *stubRoleRepo) FindByName(ctx context.Context, name string) (*models.Role, error) {
	return &models.Role{BaseModel: models.BaseModel{ID: uuid.New()}, Name: name}, nil
}

func (s *stubRoleRepo) AssignRoleToUser(ctx context.Context, uid, rid uuid.UUID) error {
	return nil
}

func (s *stubRoleRepo) GetUserRoles(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
	return []models.Role{{Name: "guest"}}, nil
}

type stubEmailVerifRepo struct {
	interfaces.EmailVerificationRepository
}

func (s *stubEmailVerifRepo) Create(ctx context.Context, v *models.EmailVerification) error {
	return nil
}

type stubRefreshTokenRepo struct {
	interfaces.RefreshTokenRepository
}

func (s *stubRefreshTokenRepo) Create(ctx context.Context, t *models.RefreshToken) error {
	return nil
}

type stubAuditLogRepo struct {
	interfaces.AuditLogRepository
}

func (s *stubAuditLogRepo) Create(ctx context.Context, l *models.AuditLog) error {
	return nil
}

func TestAuthService_SignUp(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			BCryptCost:          bcrypt.MinCost,
			VerificationTimeout: time.Hour,
		},
	}
	jwtManager := jwt.NewJWTManager("secret", "refresh", "issuer", "aud", 15*time.Minute, 24*time.Hour)
	emailService := email.NewEmailService(&config.EmailConfig{
		TemplatesPath: ".",
	})

	userRepo := &stubUserRepo{users: make(map[string]*models.User)}
	roleRepo := &stubRoleRepo{}
	emailVerifRepo := &stubEmailVerifRepo{}
	refreshTokenRepo := &stubRefreshTokenRepo{}
	auditLogRepo := &stubAuditLogRepo{}

	authSvc := NewAuthService(
		userRepo,
		roleRepo,
		nil,
		refreshTokenRepo,
		emailVerifRepo,
		auditLogRepo,
		jwtManager,
		emailService,
		cfg,
		nil, // db == nil to bypass tx creation just to unit test the flow
	)

	req := &requests.SignUpRequest{
		Email:       "test@example.com",
		Password:    "Password123!",
		FirstName:   "Test",
		LastName:    "User",
		PhoneNumber: "1234567890",
		Role:        "guest",
	}

	resp, err := authSvc.SignUp(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error during signup, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}

	if resp.User.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, resp.User.Email)
	}

	if resp.AccessToken == "" {
		t.Error("expected access token to be generated")
	}
}

func TestAuthService_SignIn(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			BCryptCost:          bcrypt.MinCost,
			VerificationTimeout: time.Hour,
		},
	}
	jwtManager := jwt.NewJWTManager("secret", "refresh", "issuer", "aud", 15*time.Minute, 24*time.Hour)

	userRepo := &stubUserRepo{users: make(map[string]*models.User)}

	// Seed user
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.MinCost)
	now := time.Now()
	userRepo.users["test@example.com"] = &models.User{
		BaseModel:       models.BaseModel{ID: uuid.New()},
		Email:           "test@example.com",
		PasswordHash:    string(hash),
		Status:          models.UserStatusActive,
		EmailVerifiedAt: &now,
	}

	authSvc := NewAuthService(
		userRepo,
		&stubRoleRepo{},
		nil,
		&stubRefreshTokenRepo{},
		&stubEmailVerifRepo{},
		&stubAuditLogRepo{},
		jwtManager,
		email.NewEmailService(&config.EmailConfig{TemplatesPath: "."}),
		cfg,
		nil,
	)

	req := &requests.SignInRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	resp, err := authSvc.SignIn(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error during signin, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email to match")
	}
}
