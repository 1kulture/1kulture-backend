package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/email"
	"github.com/1kulture/1kulture-backend/internal/utils/jwt"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type authService struct {
	userRepo          interfaces.UserRepository
	roleRepo          interfaces.RoleRepository
	refreshTokenRepo  interfaces.RefreshTokenRepository
	emailVerifRepo    interfaces.EmailVerificationRepository
	auditLogRepo      interfaces.AuditLogRepository
	passwordResetRepo interfaces.PasswordResetRepository
	jwtManager        *jwt.JWTManager
	emailService      *email.EmailService
	config            *config.Config
}

func NewAuthService(
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	passwordResetRepo interfaces.PasswordResetRepository,
	refreshTokenRepo interfaces.RefreshTokenRepository,
	emailVerifRepo interfaces.EmailVerificationRepository,
	auditLogRepo interfaces.AuditLogRepository,
	jwtManager *jwt.JWTManager,
	emailService *email.EmailService,
	cfg *config.Config,
) serviceInterfaces.AuthService {
	return &authService{
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		passwordResetRepo: passwordResetRepo,
		refreshTokenRepo:  refreshTokenRepo,
		emailVerifRepo:    emailVerifRepo,
		auditLogRepo:      auditLogRepo,
		jwtManager:        jwtManager,
		emailService:      emailService,
		config:            cfg,
	}
}

func (s *authService) SignUp(ctx context.Context, req *requests.SignUpRequest) (*responses.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.Security.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PhoneNumber:  req.PhoneNumber,
		Status:       models.UserStatusPending,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Assign default role (guest) or selected role
	roleName := string(models.RoleGuest)
	if req.Role != "" {
		roleName = req.Role
	}

	role, err := s.roleRepo.FindByName(ctx, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found: %s", roleName)
	}

	if err := s.roleRepo.AssignRoleToUser(ctx, user.ID, role.ID); err != nil {
		return nil, err
	}

	// Generate verification code
	code, err := generateVerificationCode()
	if err != nil {
		return nil, err
	}

	// Create email verification
	verification := &models.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.config.Security.VerificationTimeout),
	}

	if err := s.emailVerifRepo.Create(ctx, verification); err != nil {
		return nil, err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
		logger.Error("Failed to send verification email:", err)
		// Don't fail signup if email fails, user can resend
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "USER_SIGNUP", "user", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	// Generate tokens
	roles := []string{roleName}
	tokens, err := s.jwtManager.GenerateTokens(user.ID, user.Email, roles)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: time.Unix(tokens.RtExpires, 0),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Prepare response
	userResponse := toUserResponse(user, []models.Role{*role})

	return &responses.AuthResponse{
		User:         *userResponse,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.AtExpires,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) SignIn(ctx context.Context, req *requests.SignInRequest) (*responses.AuthResponse, error) {
	// Find user
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if account is active
	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("account is not active")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Log failed login attempt
		s.createAuditLog(ctx, &user.ID, "USER_SIGNIN_FAILED", "user", user.ID.String(), "failed")
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if email is verified
	if user.EmailVerifiedAt == nil {
		return nil, fmt.Errorf("email not verified")
	}

	// Update last login
	now := time.Now()
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		logger.Error("Failed to update last login:", err)
	}

	// Get user roles
	roles, err := s.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokens(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: time.Unix(tokens.RtExpires, 0),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "USER_SIGNIN", "user", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	// Prepare response
	userResponse := toUserResponse(user, roles)

	return &responses.AuthResponse{
		User:         *userResponse,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.AtExpires,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *requests.VerifyEmailRequest) error {
	// Find verification
	verification, err := s.emailVerifRepo.FindByEmailAndCode(ctx, req.Email, req.Code)
	if err != nil {
		return err
	}
	if verification == nil {
		return fmt.Errorf("invalid verification code")
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		return fmt.Errorf("email already verified")
	}

	// Check if expired
	if time.Now().After(verification.ExpiresAt) {
		return fmt.Errorf("verification code expired")
	}

	// Check attempts
	if verification.Attempts >= s.config.Security.MaxVerificationRetry {
		return fmt.Errorf("too many attempts, please resend code")
	}

	// Find user
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Update user email verification
	now := time.Now()
	if err := s.userRepo.UpdateEmailVerification(ctx, user.ID, now); err != nil {
		return err
	}

	// Update user status to active
	if err := s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusActive); err != nil {
		return err
	}

	// Mark verification as verified
	if err := s.emailVerifRepo.MarkAsVerified(ctx, verification.ID, now); err != nil {
		return err
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "EMAIL_VERIFIED", "user", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

func (s *authService) ResendVerification(ctx context.Context, req *requests.ResendVerificationRequest) error {
	// Find user
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Check if already verified
	if user.EmailVerifiedAt != nil {
		return fmt.Errorf("email already verified")
	}

	// Generate new code
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	// Create new verification
	verification := &models.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.config.Security.VerificationTimeout),
	}

	if err := s.emailVerifRepo.Create(ctx, verification); err != nil {
		return err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *requests.RefreshTokenRequest) (*responses.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token exists in database
	refreshToken, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil || !refreshToken.IsValid() {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("account is not active")
	}

	// Get user roles
	roles, err := s.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate new tokens
	tokens, err := s.jwtManager.GenerateTokens(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	if err := s.refreshTokenRepo.Revoke(ctx, refreshToken.ID); err != nil {
		return nil, err
	}

	// Save new refresh token
	newRefreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: time.Unix(tokens.RtExpires, 0),
	}
	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, err
	}

	return &responses.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.AtExpires,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *requests.LogoutRequest) error {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return err
	}

	// Find token
	refreshToken, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return fmt.Errorf("invalid refresh token")
	}

	// Revoke token
	if err := s.refreshTokenRepo.Revoke(ctx, refreshToken.ID); err != nil {
		return err
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &claims.UserID, "USER_LOGOUT", "auth", claims.SessionID, "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *requests.ForgotPasswordRequest) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	// Don't reveal if user exists or not (security)
	if user == nil {
		logger.WithFields(map[string]interface{}{
			"email": req.Email,
		}).Info("Password reset requested for non-existent email")
		return nil
	}

	// Generate reset token
	resetToken, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create password reset record
	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(s.config.Security.PasswordResetTimeout),
	}

	if err := s.passwordResetRepo.Create(ctx, passwordReset); err != nil {
		return err
	}

	// Create reset link
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.App.WebURL, resetToken)

	// Send password reset email
	if err := s.emailService.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		logger.Error("Failed to send password reset email:", err)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "PASSWORD_RESET_REQUESTED", "auth", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *requests.ResetPasswordRequest) error {
	// Find password reset record
	passwordReset, err := s.passwordResetRepo.FindByToken(ctx, req.Token)
	if err != nil {
		return err
	}
	if passwordReset == nil || !passwordReset.IsValid() {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, passwordReset.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.Security.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return err
	}

	// Mark reset token as used
	now := time.Now()
	if err := s.passwordResetRepo.MarkAsUsed(ctx, passwordReset.ID, now); err != nil {
		logger.Error("Failed to mark reset token as used:", err)
	}

	// Revoke all refresh tokens for this user
	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		logger.Error("Failed to revoke refresh tokens:", err)
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "PASSWORD_RESET_COMPLETED", "auth", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID string, req *requests.ChangePasswordRequest) error {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Check if new password is different from current
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.NewPassword)); err == nil {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.Security.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return err
	}

	// Revoke all refresh tokens for this user
	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		logger.Error("Failed to revoke refresh tokens:", err)
	}

	// Create audit log
	auditLog := s.createAuditLog(ctx, &user.ID, "PASSWORD_CHANGED", "auth", user.ID.String(), "success")
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

// createAuditLog is a helper function to create audit log entries
func (s *authService) createAuditLog(ctx context.Context, userID *uuid.UUID, action, resource, resourceID, status string) *models.AuditLog {
	ipAddress := "unknown"
	if ip, ok := ctx.Value("ip_address").(string); ok {
		ipAddress = ip
	}

	userAgent := "unknown"
	if ua, ok := ctx.Value("user_agent").(string); ok {
		userAgent = ua
	}

	return &models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
	}
}

// generateVerificationCode generates a 6-digit verification code
func generateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// generateSecureToken generates a secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Helper function to convert models.User to responses.UserResponse
func toUserResponse(user *models.User, roles []models.Role) *responses.UserResponse {
	roleResponses := make([]responses.RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, responses.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &responses.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		PhoneNumber:     user.PhoneNumber,
		AvatarURL:       user.AvatarURL,
		Status:          string(user.Status),
		EmailVerifiedAt: user.EmailVerifiedAt,
		LastLoginAt:     user.LastLoginAt,
		Roles:           roleResponses,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
