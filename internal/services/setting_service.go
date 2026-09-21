package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type settingService struct {
	settingRepo  interfaces.SettingRepository
	auditLogRepo interfaces.AuditLogRepository
}

func NewSettingService(settingRepo interfaces.SettingRepository, auditLogRepo interfaces.AuditLogRepository) serviceInterfaces.SettingService {
	return &settingService{settingRepo: settingRepo, auditLogRepo: auditLogRepo}
}

func (s *settingService) Create(ctx context.Context, actorID uuid.UUID, req *requests.CreateSettingRequest) (*responses.SettingResponse, error) {
	existing, err := s.settingRepo.FindByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("setting with key '%s' already exists", req.Key)
	}
	setting := &models.Setting{
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   req.ValueType,
		Category:    req.Category,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UpdatedBy:   &actorID,
	}
	if err := s.settingRepo.Create(ctx, setting); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "SETTING_CREATED", req.Key, "")
	return toSettingResponse(setting), nil
}

func (s *settingService) GetByKey(ctx context.Context, key string) (*responses.SettingResponse, error) {
	setting, err := s.settingRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, fmt.Errorf("setting not found")
	}
	return toSettingResponse(setting), nil
}

func (s *settingService) List(ctx context.Context, category string) ([]responses.SettingResponse, error) {
	settings, err := s.settingRepo.FindAll(ctx, category)
	if err != nil {
		return nil, err
	}
	out := make([]responses.SettingResponse, 0, len(settings))
	for i := range settings {
		out = append(out, *toSettingResponse(&settings[i]))
	}
	return out, nil
}

func (s *settingService) ListPublic(ctx context.Context) (*responses.PublicSettingsResponse, error) {
	settings, err := s.settingRepo.FindPublic(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(settings))
	for _, st := range settings {
		m[st.Key] = st.Value
	}
	return &responses.PublicSettingsResponse{Settings: m}, nil
}

func (s *settingService) Update(ctx context.Context, actorID uuid.UUID, req *requests.UpdateSettingRequest) (*responses.SettingResponse, error) {
	setting, err := s.settingRepo.FindByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, fmt.Errorf("setting not found")
	}
	setting.Value = req.Value
	setting.UpdatedBy = &actorID
	if err := s.settingRepo.Update(ctx, setting); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "SETTING_UPDATED", req.Key, req.Value)
	return toSettingResponse(setting), nil
}

func (s *settingService) BulkUpdate(ctx context.Context, actorID uuid.UUID, req *requests.BulkUpdateSettingsRequest) ([]responses.SettingResponse, error) {
	out := make([]responses.SettingResponse, 0, len(req.Settings))
	for _, item := range req.Settings {
		res, err := s.Update(ctx, actorID, &item)
		if err != nil {
			return nil, err
		}
		out = append(out, *res)
	}
	return out, nil
}

func (s *settingService) Delete(ctx context.Context, key string) error {
	if err := s.settingRepo.Delete(ctx, key); err != nil {
		return err
	}
	s.audit(ctx, nil, "SETTING_DELETED", key, "")
	return nil
}

func (s *settingService) GetInt(ctx context.Context, key string, fallback int) int {
	setting, err := s.settingRepo.FindByKey(ctx, key)
	if err != nil || setting == nil {
		return fallback
	}
	v, err := strconv.Atoi(setting.Value)
	if err != nil {
		return fallback
	}
	return v
}

func (s *settingService) GetString(ctx context.Context, key string, fallback string) string {
	setting, err := s.settingRepo.FindByKey(ctx, key)
	if err != nil || setting == nil {
		return fallback
	}
	return setting.Value
}

func (s *settingService) GetBool(ctx context.Context, key string, fallback bool) bool {
	setting, err := s.settingRepo.FindByKey(ctx, key)
	if err != nil || setting == nil {
		return fallback
	}
	v, err := strconv.ParseBool(setting.Value)
	if err != nil {
		return fallback
	}
	return v
}

func (s *settingService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID, details string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "setting",
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log:", err)
	}
}

func toSettingResponse(s *models.Setting) *responses.SettingResponse {
	return &responses.SettingResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       s.Value,
		ValueType:   s.ValueType,
		Category:    s.Category,
		Description: s.Description,
		IsPublic:    s.IsPublic,
		UpdatedBy:   s.UpdatedBy,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
