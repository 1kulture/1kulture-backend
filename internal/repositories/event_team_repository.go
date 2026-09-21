package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

// ---------- Co-Organizers ----------

type eventCoOrganizerRepository struct {
	db *gorm.DB
}

func NewEventCoOrganizerRepository(db *gorm.DB) interfaces.EventCoOrganizerRepository {
	return &eventCoOrganizerRepository{db: db}
}

func (r *eventCoOrganizerRepository) Create(ctx context.Context, co *models.EventCoOrganizer) error {
	if err := r.db.WithContext(ctx).Create(co).Error; err != nil {
		return fmt.Errorf("failed to create co-organizer: %w", err)
	}
	return nil
}

func (r *eventCoOrganizerRepository) FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*models.EventCoOrganizer, error) {
	var c models.EventCoOrganizer
	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&c, "event_id = ? AND user_id = ?", eventID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find co-organizer: %w", err)
	}
	return &c, nil
}

func (r *eventCoOrganizerRepository) FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventCoOrganizer, error) {
	var list []models.EventCoOrganizer
	if err := r.db.WithContext(ctx).Preload("User").Where("event_id = ?", eventID).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list co-organizers: %w", err)
	}
	return list, nil
}

func (r *eventCoOrganizerRepository) Delete(ctx context.Context, eventID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&models.EventCoOrganizer{}).Error; err != nil {
		return fmt.Errorf("failed to delete co-organizer: %w", err)
	}
	return nil
}

func (r *eventCoOrganizerRepository) IsCoOrganizer(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventCoOrganizer{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check co-organizer: %w", err)
	}
	return count > 0, nil
}

// ---------- Staff ----------

type eventStaffRepository struct {
	db *gorm.DB
}

func NewEventStaffRepository(db *gorm.DB) interfaces.EventStaffRepository {
	return &eventStaffRepository{db: db}
}

func (r *eventStaffRepository) Create(ctx context.Context, staff *models.EventStaff) error {
	if err := r.db.WithContext(ctx).Create(staff).Error; err != nil {
		return fmt.Errorf("failed to create staff: %w", err)
	}
	return nil
}

func (r *eventStaffRepository) FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*models.EventStaff, error) {
	var s models.EventStaff
	if err := r.db.WithContext(ctx).Preload("User").First(&s, "event_id = ? AND user_id = ?", eventID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find staff: %w", err)
	}
	return &s, nil
}

func (r *eventStaffRepository) FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventStaff, error) {
	var list []models.EventStaff
	if err := r.db.WithContext(ctx).Preload("User").Where("event_id = ?", eventID).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list staff: %w", err)
	}
	return list, nil
}

func (r *eventStaffRepository) Delete(ctx context.Context, eventID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&models.EventStaff{}).Error; err != nil {
		return fmt.Errorf("failed to delete staff: %w", err)
	}
	return nil
}

func (r *eventStaffRepository) IsStaff(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventStaff{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check staff: %w", err)
	}
	return count > 0, nil
}

// ---------- Event Followers ----------

type eventFollowerRepository struct {
	db *gorm.DB
}

func NewEventFollowerRepository(db *gorm.DB) interfaces.EventFollowerRepository {
	return &eventFollowerRepository{db: db}
}

func (r *eventFollowerRepository) Follow(ctx context.Context, follower *models.EventFollower) error {
	if err := r.db.WithContext(ctx).Create(follower).Error; err != nil {
		return fmt.Errorf("failed to follow event: %w", err)
	}
	return nil
}

func (r *eventFollowerRepository) Unfollow(ctx context.Context, eventID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&models.EventFollower{}).Error; err != nil {
		return fmt.Errorf("failed to unfollow event: %w", err)
	}
	return nil
}

func (r *eventFollowerRepository) IsFollowing(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventFollower{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check follower: %w", err)
	}
	return count > 0, nil
}

func (r *eventFollowerRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.EventFollower, int64, error) {
	var list []models.EventFollower
	var total int64
	q := r.db.WithContext(ctx).Model(&models.EventFollower{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count followers: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("User").Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list followers: %w", err)
	}
	return list, total, nil
}

func (r *eventFollowerRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventFollower{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count followers: %w", err)
	}
	return count, nil
}

// ---------- Organizer Followers ----------

type organizerFollowerRepository struct {
	db *gorm.DB
}

func NewOrganizerFollowerRepository(db *gorm.DB) interfaces.OrganizerFollowerRepository {
	return &organizerFollowerRepository{db: db}
}

func (r *organizerFollowerRepository) Follow(ctx context.Context, follower *models.EventOrganizerFollower) error {
	if err := r.db.WithContext(ctx).Create(follower).Error; err != nil {
		return fmt.Errorf("failed to follow organizer: %w", err)
	}
	return nil
}

func (r *organizerFollowerRepository) Unfollow(ctx context.Context, organizerID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("organizer_id = ? AND user_id = ?", organizerID, userID).Delete(&models.EventOrganizerFollower{}).Error; err != nil {
		return fmt.Errorf("failed to unfollow organizer: %w", err)
	}
	return nil
}

func (r *organizerFollowerRepository) IsFollowing(ctx context.Context, organizerID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventOrganizerFollower{}).
		Where("organizer_id = ? AND user_id = ?", organizerID, userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check organizer follower: %w", err)
	}
	return count > 0, nil
}

func (r *organizerFollowerRepository) CountByOrganizer(ctx context.Context, organizerID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventOrganizerFollower{}).Where("organizer_id = ?", organizerID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count organizer followers: %w", err)
	}
	return count, nil
}

func (r *organizerFollowerRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID, page, perPage int) ([]models.EventOrganizerFollower, int64, error) {
	var list []models.EventOrganizerFollower
	var total int64
	q := r.db.WithContext(ctx).Model(&models.EventOrganizerFollower{}).Where("organizer_id = ?", organizerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list: %w", err)
	}
	return list, total, nil
}

// ---------- Event Shares ----------

type eventShareRepository struct {
	db *gorm.DB
}

func NewEventShareRepository(db *gorm.DB) interfaces.EventShareRepository {
	return &eventShareRepository{db: db}
}

func (r *eventShareRepository) Create(ctx context.Context, share *models.EventShare) error {
	if err := r.db.WithContext(ctx).Create(share).Error; err != nil {
		return fmt.Errorf("failed to record share: %w", err)
	}
	return nil
}

func (r *eventShareRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EventShare{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count shares: %w", err)
	}
	return count, nil
}

func (r *eventShareRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.EventShare, int64, error) {
	var list []models.EventShare
	var total int64
	q := r.db.WithContext(ctx).Model(&models.EventShare{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list: %w", err)
	}
	return list, total, nil
}
