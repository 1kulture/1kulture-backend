package interfaces

import (
	"context"

	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type EventService interface {
	// ---------- CRUD ----------
	Create(ctx context.Context, actorID uuid.UUID, req *requests.EventCreateRequest) (*responses.EventResponse, error)
	GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*responses.EventResponse, error)
	GetBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*responses.EventResponse, error)
	Update(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventUpdateRequest) (*responses.EventResponse, error)
	Delete(ctx context.Context, actorID, eventID uuid.UUID) error
	List(ctx context.Context, filter repoInterfaces.EventListFilter, viewerID *uuid.UUID) ([]responses.EventSummaryResponse, int64, error)

	// ---------- Lifecycle ----------
	Publish(ctx context.Context, actorID, eventID uuid.UUID) (*responses.EventResponse, error)
	Unpublish(ctx context.Context, actorID, eventID uuid.UUID) (*responses.EventResponse, error)
	Cancel(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventCancelRequest) (*responses.EventResponse, error)
	Postpone(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventPostponeRequest) (*responses.EventResponse, error)

	// ---------- Occurrences ----------
	AddOccurrence(ctx context.Context, actorID, eventID uuid.UUID, req *requests.OccurrenceCreateRequest) (*responses.EventOccurrenceResponse, error)
	UpdateOccurrence(ctx context.Context, actorID, eventID, occID uuid.UUID, req *requests.OccurrenceUpdateRequest) (*responses.EventOccurrenceResponse, error)
	DeleteOccurrence(ctx context.Context, actorID, eventID, occID uuid.UUID) error
	ListOccurrences(ctx context.Context, eventID uuid.UUID) ([]responses.EventOccurrenceResponse, error)

	// ---------- Team: Co-Organizers ----------
	AddCoOrganizer(ctx context.Context, actorID, eventID uuid.UUID, req *requests.CoOrganizerAddRequest) (*responses.EventCoOrganizerResponse, error)
	RemoveCoOrganizer(ctx context.Context, actorID, eventID, targetUserID uuid.UUID) error
	ListCoOrganizers(ctx context.Context, eventID uuid.UUID) ([]responses.EventCoOrganizerResponse, error)

	// ---------- Team: Staff ----------
	AddStaff(ctx context.Context, actorID, eventID uuid.UUID, req *requests.StaffAddRequest) (*responses.EventStaffResponse, error)
	RemoveStaff(ctx context.Context, actorID, eventID, targetUserID uuid.UUID) error
	ListStaff(ctx context.Context, eventID uuid.UUID) ([]responses.EventStaffResponse, error)

	// ---------- Follow ----------
	FollowEvent(ctx context.Context, userID, eventID uuid.UUID) error
	UnfollowEvent(ctx context.Context, userID, eventID uuid.UUID) error
	ListEventFollowers(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.EventFollowerResponse, int64, error)

	FollowOrganizer(ctx context.Context, userID, organizerID uuid.UUID) error
	UnfollowOrganizer(ctx context.Context, userID, organizerID uuid.UUID) error

	// ---------- Share ----------
	RecordShare(ctx context.Context, eventID uuid.UUID, userID *uuid.UUID, req *requests.ShareRequest, ip, userAgent string) (*responses.EventShareResponse, error)
}
