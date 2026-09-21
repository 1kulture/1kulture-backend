package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type EventCoOrganizerRepository interface {
	Create(ctx context.Context, co *models.EventCoOrganizer) error
	FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*models.EventCoOrganizer, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventCoOrganizer, error)
	Delete(ctx context.Context, eventID, userID uuid.UUID) error
	IsCoOrganizer(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

type EventStaffRepository interface {
	Create(ctx context.Context, staff *models.EventStaff) error
	FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*models.EventStaff, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventStaff, error)
	Delete(ctx context.Context, eventID, userID uuid.UUID) error
	IsStaff(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

type EventFollowerRepository interface {
	Follow(ctx context.Context, follower *models.EventFollower) error
	Unfollow(ctx context.Context, eventID, userID uuid.UUID) error
	IsFollowing(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.EventFollower, int64, error)
	CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error)
}

type OrganizerFollowerRepository interface {
	Follow(ctx context.Context, follower *models.EventOrganizerFollower) error
	Unfollow(ctx context.Context, organizerID, userID uuid.UUID) error
	IsFollowing(ctx context.Context, organizerID, userID uuid.UUID) (bool, error)
	CountByOrganizer(ctx context.Context, organizerID uuid.UUID) (int64, error)
	FindByOrganizer(ctx context.Context, organizerID uuid.UUID, page, perPage int) ([]models.EventOrganizerFollower, int64, error)
}

type EventShareRepository interface {
	Create(ctx context.Context, share *models.EventShare) error
	CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.EventShare, int64, error)
}
