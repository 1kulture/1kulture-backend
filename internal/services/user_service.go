package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type userService struct {
	userRepo     interfaces.UserRepository
	roleRepo     interfaces.RoleRepository
	auditLogRepo interfaces.AuditLogRepository
	kycRepo      interfaces.KYCRepository
}

func NewUserService(
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	auditLogRepo interfaces.AuditLogRepository,
	kycRepo interfaces.KYCRepository,
) serviceInterfaces.UserService {
	return &userService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditLogRepo: auditLogRepo,
		kycRepo:      kycRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*responses.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return toUserResponse(user, user.Roles), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *requests.UpdateProfileRequest) (*responses.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Get updated user with roles
	updatedUser, err := s.userRepo.WithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     "PROFILE_UPDATED",
		Resource:   "user",
		ResourceID: userID.String(),
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return toUserResponse(updatedUser, updatedUser.Roles), nil
}

func (s *userService) UpdateRole(ctx context.Context, userID uuid.UUID, req *requests.UpdateRoleRequest) error {
	// Find the role
	role, err := s.roleRepo.FindByName(ctx, req.Role)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Assign role to user
	if err := s.roleRepo.AssignRoleToUser(ctx, userID, role.ID); err != nil {
		return err
	}

	// Create audit log
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     "ROLE_UPDATED",
		Resource:   "user",
		ResourceID: userID.String(),
		Details:    fmt.Sprintf("Role added: %s", req.Role),
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return nil
}

func (s *userService) SubmitKYC(ctx context.Context, userID uuid.UUID, req *requests.KYCRequest) (*responses.KYCResponse, error) {
	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID")
	}

	// Verify role exists
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Check if user already has pending KYC
	existingKYC, err := s.kycRepo.FindByUserAndRole(ctx, userID, roleID)
	if err != nil {
		return nil, err
	}
	if existingKYC != nil && existingKYC.Status == models.KYCStatusPending {
		return nil, fmt.Errorf("KYC verification already pending for this role")
	}

	// Create KYC verification
	kyc := &models.KYCVerification{
		UserID:            userID,
		RoleID:            roleID,
		Status:            models.KYCStatusPending,
		DocumentType:      req.DocumentType,
		DocumentNumber:    req.DocumentNumber,
		DocumentURL:       req.DocumentURL,
		SelfieURL:         req.SelfieURL,
		AddressProofURL:   req.AddressProofURL,
		BusinessName:      req.BusinessName,
		BusinessRegNumber: req.BusinessRegNumber,
		SubmittedAt:       time.Now(),
	}

	if err := s.kycRepo.Create(ctx, kyc); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     "KYC_SUBMITTED",
		Resource:   "kyc",
		ResourceID: kyc.ID.String(),
		Details:    fmt.Sprintf("KYC submitted for role: %s", role.Name),
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log:", err)
	}

	return toKYCResponse(kyc, role.Name), nil
}

func (s *userService) GetKYCStatus(ctx context.Context, userID uuid.UUID) (*responses.KYCResponse, error) {
	// Get latest KYC submission
	kyc, err := s.kycRepo.FindLatestByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if kyc == nil {
		return nil, fmt.Errorf("no KYC submission found")
	}

	// Get role name
	role, err := s.roleRepo.FindByID(ctx, kyc.RoleID)
	if err != nil {
		return nil, err
	}
	roleName := ""
	if role != nil {
		roleName = role.Name
	}

	return toKYCResponse(kyc, roleName), nil
}

func toKYCResponse(kyc *models.KYCVerification, roleName string) *responses.KYCResponse {
	return &responses.KYCResponse{
		ID:                kyc.ID,
		UserID:            kyc.UserID,
		RoleID:            kyc.RoleID,
		RoleName:          roleName,
		Status:            string(kyc.Status),
		DocumentType:      kyc.DocumentType,
		DocumentNumber:    kyc.DocumentNumber,
		DocumentURL:       kyc.DocumentURL,
		SelfieURL:         kyc.SelfieURL,
		AddressProofURL:   kyc.AddressProofURL,
		BusinessName:      kyc.BusinessName,
		BusinessRegNumber: kyc.BusinessRegNumber,
		SubmittedAt:       kyc.SubmittedAt,
		ReviewedAt:        kyc.ReviewedAt,
		RejectionReason:   kyc.RejectionReason,
	}
}

func getContextString(ctx context.Context, key string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	return ""
}
