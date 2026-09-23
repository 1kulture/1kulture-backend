package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type brandProfileService struct {
	brandRepo       repoInterfaces.BrandProfileRepository
	roleRepo        repoInterfaces.RoleRepository
	userRepo        repoInterfaces.UserRepository
	partnershipRepo repoInterfaces.PartnershipRequestRepository
	auditLogRepo    repoInterfaces.AuditLogRepository
}

func NewBrandProfileService(
	brandRepo repoInterfaces.BrandProfileRepository,
	roleRepo repoInterfaces.RoleRepository,
	userRepo repoInterfaces.UserRepository,
	partnershipRepo repoInterfaces.PartnershipRequestRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.BrandProfileService {
	return &brandProfileService{
		brandRepo:       brandRepo,
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		partnershipRepo: partnershipRepo,
		auditLogRepo:    auditLogRepo,
	}
}

func (s *brandProfileService) CreateForUser(ctx context.Context, userID uuid.UUID, req *requests.CreateBrandProfileRequest) (*responses.BrandProfileResponse, error) {
	existing, err := s.brandRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("brand profile already exists for this user")
	}

	preferredJSON := toJSON(req.PreferredCategories)
	tagsJSON := toJSON(req.Tags)

	size := models.BrandSize(req.BusinessSize)
	if size == "" {
		size = models.BrandSizeSmall
	}

	profile := &models.BrandProfile{
		UserID:              userID,
		BusinessName:        req.BusinessName,
		LogoURL:             req.LogoURL,
		Industry:            req.Industry,
		Description:         req.Description,
		Website:             req.Website,
		Instagram:           req.Instagram,
		Location:            req.Location,
		ContactName:         req.ContactName,
		ContactEmail:        req.ContactEmail,
		ContactPhone:        req.ContactPhone,
		BusinessSize:        size,
		TargetAudience:      req.TargetAudience,
		PreferredCategories: preferredJSON,
		Tags:                tagsJSON,
		IsActive:            true,
	}
	if err := s.brandRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	// Auto-assign brand role
	if role, err := s.roleRepo.FindByName(ctx, string(models.RoleBrand)); err == nil && role != nil {
		if err := s.roleRepo.AssignRoleToUser(ctx, userID, role.ID); err != nil {
			logger.Error("Failed to assign brand role: ", err)
		}
	}

	s.audit(ctx, &userID, "BRAND_PROFILE_CREATED", profile.ID.String())
	return toBrandProfileResponse(profile), nil
}

func (s *brandProfileService) GetMine(ctx context.Context, userID uuid.UUID) (*responses.BrandProfileResponse, error) {
	profile, err := s.brandRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("brand profile not found")
	}
	res := toBrandProfileResponse(profile)

	// Aggregated counts
	if counts, err := s.partnershipRepo.CountByStatus(ctx, repoInterfaces.PartnershipRequestFilter{BrandUserID: &userID}); err == nil {
		res.ActivePartnerships = counts[string(models.PartnershipStatusActive)] + counts[string(models.PartnershipStatusAccepted)]
		total := 0
		for _, c := range counts {
			total += c
		}
		res.TotalPartnerships = total
	}
	return res, nil
}

func (s *brandProfileService) GetByID(ctx context.Context, id uuid.UUID) (*responses.BrandProfileResponse, error) {
	profile, err := s.brandRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("brand profile not found")
	}
	return toBrandProfileResponse(profile), nil
}

func (s *brandProfileService) GetByUserID(ctx context.Context, userID uuid.UUID) (*responses.BrandProfileResponse, error) {
	profile, err := s.brandRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("brand profile not found")
	}
	return toBrandProfileResponse(profile), nil
}

func (s *brandProfileService) Update(ctx context.Context, userID uuid.UUID, req *requests.UpdateBrandProfileRequest) (*responses.BrandProfileResponse, error) {
	profile, err := s.brandRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("brand profile not found")
	}

	if req.BusinessName != nil {
		profile.BusinessName = *req.BusinessName
	}
	if req.LogoURL != nil {
		profile.LogoURL = *req.LogoURL
	}
	if req.Industry != nil {
		profile.Industry = *req.Industry
	}
	if req.Description != nil {
		profile.Description = *req.Description
	}
	if req.Website != nil {
		profile.Website = *req.Website
	}
	if req.Instagram != nil {
		profile.Instagram = *req.Instagram
	}
	if req.Location != nil {
		profile.Location = *req.Location
	}
	if req.ContactName != nil {
		profile.ContactName = *req.ContactName
	}
	if req.ContactEmail != nil {
		profile.ContactEmail = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		profile.ContactPhone = *req.ContactPhone
	}
	if req.BusinessSize != nil {
		profile.BusinessSize = models.BrandSize(*req.BusinessSize)
	}
	if req.TargetAudience != nil {
		profile.TargetAudience = *req.TargetAudience
	}
	if req.PreferredCategories != nil {
		profile.PreferredCategories = toJSON(req.PreferredCategories)
	}
	if req.Tags != nil {
		profile.Tags = toJSON(req.Tags)
	}
	if req.IsActive != nil {
		profile.IsActive = *req.IsActive
	}

	if err := s.brandRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	s.audit(ctx, &userID, "BRAND_PROFILE_UPDATED", profile.ID.String())
	return toBrandProfileResponse(profile), nil
}

func (s *brandProfileService) List(ctx context.Context, industry, location string, page, perPage int) ([]responses.BrandProfileResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	list, total, err := s.brandRepo.List(ctx, industry, location, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.BrandProfileResponse, 0, len(list))
	for i := range list {
		out = append(out, *toBrandProfileResponse(&list[i]))
	}
	return out, total, nil
}

func (s *brandProfileService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "brand_profile",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toBrandProfileResponse(p *models.BrandProfile) *responses.BrandProfileResponse {
	preferred := fromJSONStrings(p.PreferredCategories)
	tags := fromJSONStrings(p.Tags)
	return &responses.BrandProfileResponse{
		ID:                  p.ID,
		UserID:              p.UserID,
		BusinessName:        p.BusinessName,
		LogoURL:             p.LogoURL,
		Industry:            p.Industry,
		Description:         p.Description,
		Website:             p.Website,
		Instagram:           p.Instagram,
		Location:            p.Location,
		ContactName:         p.ContactName,
		ContactEmail:        p.ContactEmail,
		ContactPhone:        p.ContactPhone,
		BusinessSize:        string(p.BusinessSize),
		TargetAudience:      p.TargetAudience,
		PreferredCategories: preferred,
		Tags:                tags,
		IsActive:            p.IsActive,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}

func toBrandSummary(p *models.BrandProfile) *responses.BrandSummary {
	if p == nil {
		return nil
	}
	return &responses.BrandSummary{
		ID:           p.ID,
		BusinessName: p.BusinessName,
		LogoURL:      p.LogoURL,
		Industry:     p.Industry,
		Location:     p.Location,
	}
}

// helper used only in this file to keep imports tidy
var _ = datatypes.JSON{}
