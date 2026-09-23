package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type partnershipRequestRepository struct {
	db *gorm.DB
}

func NewPartnershipRequestRepository(db *gorm.DB) repoInterfaces.PartnershipRequestRepository {
	return &partnershipRequestRepository{db: db}
}

func (r *partnershipRequestRepository) Create(ctx context.Context, req *models.PartnershipRequest) error {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return fmt.Errorf("failed to create partnership request: %w", err)
	}
	return nil
}

func (r *partnershipRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipRequest, error) {
	var req models.PartnershipRequest
	if err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find partnership request: %w", err)
	}
	return &req, nil
}

func (r *partnershipRequestRepository) FindWithRelations(ctx context.Context, id uuid.UUID) (*models.PartnershipRequest, error) {
	var req models.PartnershipRequest
	if err := r.db.WithContext(ctx).
		Preload("BrandProfile").
		Preload("Event").
		First(&req, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find partnership request with relations: %w", err)
	}
	return &req, nil
}

func (r *partnershipRequestRepository) Update(ctx context.Context, req *models.PartnershipRequest) error {
	if err := r.db.WithContext(ctx).Save(req).Error; err != nil {
		return fmt.Errorf("failed to update partnership request: %w", err)
	}
	return nil
}

func (r *partnershipRequestRepository) List(ctx context.Context, filter repoInterfaces.PartnershipRequestFilter) ([]models.PartnershipRequest, int64, error) {
	var list []models.PartnershipRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&models.PartnershipRequest{})

	if filter.BrandProfileID != nil {
		q = q.Where("brand_profile_id = ?", *filter.BrandProfileID)
	}
	if filter.BrandUserID != nil {
		q = q.Where("brand_user_id = ?", *filter.BrandUserID)
	}
	if filter.EventID != nil {
		q = q.Where("event_id = ?", *filter.EventID)
	}
	if filter.OrganizerID != nil {
		q = q.Where("organizer_id = ?", *filter.OrganizerID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if len(filter.Statuses) > 0 {
		q = q.Where("status IN ?", filter.Statuses)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count partnership requests: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	if err := q.Preload("BrandProfile").
		Preload("Event").
		Order("created_at desc").
		Limit(perPage).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list partnership requests: %w", err)
	}
	return list, total, nil
}

func (r *partnershipRequestRepository) CountByStatus(ctx context.Context, filter repoInterfaces.PartnershipRequestFilter) (map[string]int, error) {
	type row struct {
		Status string
		Count  int
	}
	var rows []row

	q := r.db.WithContext(ctx).Model(&models.PartnershipRequest{}).
		Select("status, COUNT(*) as count").
		Group("status")

	if filter.BrandProfileID != nil {
		q = q.Where("brand_profile_id = ?", *filter.BrandProfileID)
	}
	if filter.BrandUserID != nil {
		q = q.Where("brand_user_id = ?", *filter.BrandUserID)
	}
	if filter.EventID != nil {
		q = q.Where("event_id = ?", *filter.EventID)
	}
	if filter.OrganizerID != nil {
		q = q.Where("organizer_id = ?", *filter.OrganizerID)
	}

	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	out := make(map[string]int, len(rows))
	for _, rw := range rows {
		out[rw.Status] = rw.Count
	}
	return out, nil
}

func (r *partnershipRequestRepository) ExistsActive(ctx context.Context, brandUserID, eventID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.PartnershipRequest{}).
		Where("brand_user_id = ? AND event_id = ? AND status IN ?",
			brandUserID, eventID,
			[]string{string(models.PartnershipStatusPending), string(models.PartnershipStatusAccepted), string(models.PartnershipStatusActive)},
		).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check active partnership: %w", err)
	}
	return count > 0, nil
}
