package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) interfaces.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

func (r *eventRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var e models.Event
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find event: %w", err)
	}
	return &e, nil
}

func (r *eventRepository) FindBySlug(ctx context.Context, slug string) (*models.Event, error) {
	var e models.Event
	if err := r.db.WithContext(ctx).First(&e, "slug = ?", slug).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find event by slug: %w", err)
	}
	return &e, nil
}

func (r *eventRepository) FindWithRelations(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var e models.Event
	if err := r.db.WithContext(ctx).
		Preload("Occurrences", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc, start_at asc") }).
		Preload("CoOrganizers").
		Preload("Organizer").
		First(&e, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find event with relations: %w", err)
	}
	return &e, nil
}

func (r *eventRepository) Update(ctx context.Context, event *models.Event) error {
	if err := r.db.WithContext(ctx).Save(event).Error; err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Event{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

func (r *eventRepository) List(ctx context.Context, filter interfaces.EventListFilter) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Event{})

	if filter.OrganizerID != nil {
		q = q.Where("organizer_id = ?", *filter.OrganizerID)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.City != "" {
		q = q.Where("LOWER(venue_city) = LOWER(?)", filter.City)
	}
	if filter.Country != "" {
		q = q.Where("LOWER(venue_country) = LOWER(?)", filter.Country)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	} else {
		// default: only show published for public listings
		q = q.Where("status = ?", string(models.EventStatusPublished))
	}
	if filter.EventType != "" {
		q = q.Where("event_type = ?", filter.EventType)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("(LOWER(title) LIKE LOWER(?) OR LOWER(summary) LIKE LOWER(?))", like, like)
	}
	if filter.StartFrom != nil {
		q = q.Where("start_at >= ?", *filter.StartFrom)
	}
	if filter.StartTo != nil {
		q = q.Where("start_at <= ?", *filter.StartTo)
	}
	if filter.Featured != nil {
		q = q.Where("is_featured = ?", *filter.Featured)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
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

	if err := q.Preload("Organizer").
		Order("start_at asc").
		Limit(perPage).
		Offset(offset).
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}

	return events, total, nil
}

func (r *eventRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Event{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}
	return count > 0, nil
}

func (r *eventRepository) IncrementField(ctx context.Context, id uuid.UUID, field string, delta int) error {
	q := fmt.Sprintf("%s = COALESCE(%s, 0) + ?", field, field)
	if err := r.db.WithContext(ctx).Model(&models.Event{}).Where("id = ?", id).UpdateColumn(field, gorm.Expr(q, delta)).Error; err != nil {
		return fmt.Errorf("failed to increment %s: %w", field, err)
	}
	return nil
}
