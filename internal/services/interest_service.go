package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type interestService struct {
	categoryRepo repoInterfaces.EventCategoryRepository
	interestRepo repoInterfaces.UserInterestRepository
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewInterestService(
	categoryRepo repoInterfaces.EventCategoryRepository,
	interestRepo repoInterfaces.UserInterestRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.InterestService {
	return &interestService{
		categoryRepo: categoryRepo,
		interestRepo: interestRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *interestService) ListCategories(ctx context.Context) ([]responses.EventCategoryResponse, error) {
	cats, err := s.categoryRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]responses.EventCategoryResponse, 0, len(cats))
	for i := range cats {
		out = append(out, *toCategoryResponse(&cats[i]))
	}
	return out, nil
}

func (s *interestService) GetMyInterests(ctx context.Context, userID uuid.UUID) (*responses.UserInterestsResponse, error) {
	rows, err := s.interestRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.EventCategoryResponse, 0, len(rows))
	for i := range rows {
		out = append(out, *toCategoryResponse(&rows[i].Category))
	}
	return &responses.UserInterestsResponse{
		Interests: out,
		Count:     len(out),
	}, nil
}

func (s *interestService) UpdateMyInterests(ctx context.Context, userID uuid.UUID, req *requests.UpdateInterestsRequest) (*responses.UserInterestsResponse, error) {
	// Validate slugs exist
	var categoryIDs []uuid.UUID
	if len(req.CategorySlugs) > 0 {
		cats, err := s.categoryRepo.FindBySlugs(ctx, req.CategorySlugs)
		if err != nil {
			return nil, err
		}
		if len(cats) != len(req.CategorySlugs) {
			// Some slugs were invalid — build a helpful error
			found := make(map[string]bool, len(cats))
			for _, c := range cats {
				found[c.Slug] = true
			}
			var missing []string
			for _, s := range req.CategorySlugs {
				if !found[s] {
					missing = append(missing, s)
				}
			}
			return nil, fmt.Errorf("invalid categories: %v", missing)
		}
		for _, c := range cats {
			categoryIDs = append(categoryIDs, c.ID)
		}
	}

	if err := s.interestRepo.ReplaceForUser(ctx, userID, categoryIDs); err != nil {
		return nil, err
	}

	s.audit(ctx, &userID, "USER_INTERESTS_UPDATED")

	return s.GetMyInterests(ctx, userID)
}

func (s *interestService) GetInterestSlugs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	ids, err := s.interestRepo.ListCategoryIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// Fetch slugs
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		c, err := s.categoryRepo.FindByID(ctx, id)
		if err != nil || c == nil {
			continue
		}
		out = append(out, c.Slug)
	}
	return out, nil
}

func (s *interestService) audit(ctx context.Context, actorID *uuid.UUID, action string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "user_interest",
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toCategoryResponse(c *models.EventCategory) *responses.EventCategoryResponse {
	return &responses.EventCategoryResponse{
		ID:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		Icon:      c.Icon,
		SortOrder: c.SortOrder,
	}
}