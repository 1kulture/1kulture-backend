package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type brandDiscoveryService struct {
	eventRepo   repoInterfaces.EventRepository
	configRepo  repoInterfaces.EventPartnershipConfigRepository
	oppRepo     repoInterfaces.PartnershipOpportunityRepository
	brandRepo   repoInterfaces.BrandProfileRepository
	requestRepo repoInterfaces.PartnershipRequestRepository
	coOrgRepo   repoInterfaces.EventCoOrganizerRepository
}

func NewBrandDiscoveryService(
	eventRepo repoInterfaces.EventRepository,
	configRepo repoInterfaces.EventPartnershipConfigRepository,
	oppRepo repoInterfaces.PartnershipOpportunityRepository,
	brandRepo repoInterfaces.BrandProfileRepository,
	requestRepo repoInterfaces.PartnershipRequestRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
) serviceInterfaces.BrandDiscoveryService {
	return &brandDiscoveryService{
		eventRepo:   eventRepo,
		configRepo:  configRepo,
		oppRepo:     oppRepo,
		brandRepo:   brandRepo,
		requestRepo: requestRepo,
		coOrgRepo:   coOrgRepo,
	}
}

// DiscoverEvents lists published events that are open to brand partnerships.
// NOTE: audience-range and partnership-type filtering is done in-memory
// (post-query) since the config lives in a separate table. For MVP scale this
// is fine; for high volume we'd add a materialized view or denormalized column.
func (s *brandDiscoveryService) DiscoverEvents(ctx context.Context, brandUserID uuid.UUID, filter serviceInterfaces.BrandDiscoveryFilter) ([]responses.EventSummaryResponse, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Only published events, and we fetch a wider set then filter by config.
	eventFilter := repoInterfaces.EventListFilter{
		Category: filter.Category,
		City:     filter.City,
		Country:  filter.Country,
		Status:   string(models.EventStatusPublished),
		Search:   filter.Search,
		Page:     1,
		PerPage:  500, // wide fetch to allow post-filtering
	}
	if filter.StartFromISO != "" {
		if t, err := time.Parse(time.RFC3339, filter.StartFromISO); err == nil {
			eventFilter.StartFrom = &t
		}
	}
	if filter.StartToISO != "" {
		if t, err := time.Parse(time.RFC3339, filter.StartToISO); err == nil {
			eventFilter.StartTo = &t
		}
	}

	events, _, err := s.eventRepo.List(ctx, eventFilter)
	if err != nil {
		return nil, 0, err
	}

	// Post-filter: keep only events with an open partnership config matching filters
	matched := make([]models.Event, 0, len(events))
	for i := range events {
		cfg, err := s.configRepo.FindByEventID(ctx, events[i].ID)
		if err != nil || cfg == nil || !cfg.IsOpen {
			continue
		}
		if filter.MinAudience > 0 && cfg.ExpectedAttendees < filter.MinAudience {
			continue
		}
		if filter.MaxAudience > 0 && cfg.ExpectedAttendees > filter.MaxAudience {
			continue
		}
		if filter.PartnershipType != "" {
			types := fromJSONStrings(cfg.PartnershipTypes)
			found := false
			for _, t := range types {
				if t == filter.PartnershipType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		matched = append(matched, events[i])
	}

	total := int64(len(matched))
	start := (page - 1) * perPage
	if start >= len(matched) {
		return []responses.EventSummaryResponse{}, total, nil
	}
	end := start + perPage
	if end > len(matched) {
		end = len(matched)
	}

	out := make([]responses.EventSummaryResponse, 0, end-start)
	for i := start; i < end; i++ {
		out = append(out, *toEventSummary(&matched[i]))
	}
	return out, total, nil
}

func (s *brandDiscoveryService) GetEventDetail(ctx context.Context, brandUserID, eventID uuid.UUID) (*responses.EventResponse, *responses.EventPartnershipConfigResponse, error) {
	event, err := s.eventRepo.FindWithRelations(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	if event == nil {
		return nil, nil, fmt.Errorf("event not found")
	}
	if event.Status != models.EventStatusPublished {
		return nil, nil, fmt.Errorf("event is not published")
	}

	// Build a lightweight event response (reuse existing builder from event service? we don't have one exported)
	eventResp := &responses.EventResponse{
		ID:           event.ID,
		OrganizerID:  event.OrganizerID,
		Slug:         event.Slug,
		Title:        event.Title,
		Summary:      event.Summary,
		Description:  event.Description,
		BannerURL:    event.BannerURL,
		ThumbnailURL: event.ThumbnailURL,
		Gallery:      fromJSONStrings(event.Gallery),
		Category:     event.Category,
		Tags:         fromJSONStrings(event.Tags),
		EventType:    string(event.EventType),
		VenueName:    event.VenueName,
		VenueCity:    event.VenueCity,
		VenueCountry: event.VenueCountry,
		Timezone:     event.Timezone,
		StartAt:      event.StartAt,
		EndAt:        event.EndAt,
		Status:       string(event.Status),
		Visibility:   string(event.Visibility),
		Currency:     event.Currency,
		CreatedAt:    event.CreatedAt,
		UpdatedAt:    event.UpdatedAt,
	}
	if event.Organizer.ID != uuid.Nil {
		eventResp.Organizer = &responses.EventOrganizerBrief{
			ID:        event.Organizer.ID,
			FirstName: event.Organizer.FirstName,
			LastName:  event.Organizer.LastName,
			AvatarURL: event.Organizer.AvatarURL,
		}
	}

	cfg, err := s.configRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	if cfg == nil || !cfg.IsOpen {
		return eventResp, nil, nil
	}

	opps, _ := s.oppRepo.FindByEvent(ctx, eventID, false)
	cfgResp := toConfigResponse(cfg, opps)

	return eventResp, cfgResp, nil
}

func (s *brandDiscoveryService) Recommended(ctx context.Context, brandUserID uuid.UUID, limit int) ([]responses.EventSummaryResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	brand, err := s.brandRepo.FindByUserID(ctx, brandUserID)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		// Fall back to latest published events
		events, _, err := s.eventRepo.List(ctx, repoInterfaces.EventListFilter{
			Status:  string(models.EventStatusPublished),
			Page:    1,
			PerPage: limit,
		})
		if err != nil {
			return nil, err
		}
		out := make([]responses.EventSummaryResponse, 0, len(events))
		for i := range events {
			out = append(out, *toEventSummary(&events[i]))
		}
		return out, nil
	}

	preferred := fromJSONStrings(brand.PreferredCategories)
	eventFilter := repoInterfaces.EventListFilter{
		Status:  string(models.EventStatusPublished),
		Page:    1,
		PerPage: limit * 3,
	}
	if len(preferred) > 0 {
		// Query first preferred category for simple MVP matching.
		eventFilter.Category = preferred[0]
	}
	events, _, err := s.eventRepo.List(ctx, eventFilter)
	if err != nil {
		return nil, err
	}

	out := make([]responses.EventSummaryResponse, 0, len(events))
	for i := range events {
		if len(out) >= limit {
			break
		}
		cfg, err := s.configRepo.FindByEventID(ctx, events[i].ID)
		if err != nil || cfg == nil || !cfg.IsOpen {
			continue
		}
		out = append(out, *toEventSummary(&events[i]))
	}

	// Fallback if nothing matched
	if len(out) == 0 {
		fallback, _, _ := s.eventRepo.List(ctx, repoInterfaces.EventListFilter{
			Status:  string(models.EventStatusPublished),
			Page:    1,
			PerPage: limit,
		})
		for i := range fallback {
			out = append(out, *toEventSummary(&fallback[i]))
		}
	}
	return out, nil
}

func (s *brandDiscoveryService) Dashboard(ctx context.Context, brandUserID uuid.UUID) (*responses.BrandDashboardResponse, error) {
	// Active partnerships
	activeCounts, err := s.requestRepo.CountByStatus(ctx, repoInterfaces.PartnershipRequestFilter{BrandUserID: &brandUserID})
	if err != nil {
		logger.Error("Dashboard counts: ", err)
	}
	active := activeCounts[string(models.PartnershipStatusAccepted)] + activeCounts[string(models.PartnershipStatusActive)]
	pending := activeCounts[string(models.PartnershipStatusPending)]

	// Available events (open + published) — approximate via count of published events
	// A more precise count would join with the config table.
	_, totalEvents, err := s.eventRepo.List(ctx, repoInterfaces.EventListFilter{
		Status:  string(models.EventStatusPublished),
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		logger.Error("Dashboard events: ", err)
	}

	return &responses.BrandDashboardResponse{
		ActivePartnerships: active,
		PendingRequests:    pending,
		AvailableEvents:    int(totalEvents),
	}, nil
}
