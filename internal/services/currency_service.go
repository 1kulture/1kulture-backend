package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type currencyService struct {
	currencyRepo    interfaces.CurrencyRepository
	currencyReqRepo interfaces.CurrencyRequestRepository
	auditLogRepo    interfaces.AuditLogRepository
}

func NewCurrencyService(
	currencyRepo interfaces.CurrencyRepository,
	currencyReqRepo interfaces.CurrencyRequestRepository,
	auditLogRepo interfaces.AuditLogRepository,
) serviceInterfaces.CurrencyService {
	return &currencyService{
		currencyRepo:    currencyRepo,
		currencyReqRepo: currencyReqRepo,
		auditLogRepo:    auditLogRepo,
	}
}

func (s *currencyService) Create(ctx context.Context, actorID uuid.UUID, req *requests.CreateCurrencyRequest) (*responses.CurrencyResponse, error) {
	existing, err := s.currencyRepo.FindByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("currency with code '%s' already exists", req.Code)
	}
	c := &models.Currency{
		Code:           req.Code,
		Name:           req.Name,
		Symbol:         req.Symbol,
		IsEnabled:      req.IsEnabled,
		IsDefault:      req.IsDefault,
		MinAmountMinor: req.MinAmountMinor,
	}
	if err := s.currencyRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	if req.IsDefault {
		_ = s.currencyRepo.SetDefault(ctx, req.Code)
	}
	s.audit(ctx, &actorID, "CURRENCY_CREATED", req.Code)
	return toCurrencyResponse(c), nil
}

func (s *currencyService) GetByCode(ctx context.Context, code string) (*responses.CurrencyResponse, error) {
	c, err := s.currencyRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("currency not found")
	}
	return toCurrencyResponse(c), nil
}

func (s *currencyService) List(ctx context.Context, onlyEnabled bool) ([]responses.CurrencyResponse, error) {
	list, err := s.currencyRepo.FindAll(ctx, onlyEnabled)
	if err != nil {
		return nil, err
	}
	out := make([]responses.CurrencyResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCurrencyResponse(&list[i]))
	}
	return out, nil
}

func (s *currencyService) Update(ctx context.Context, code string, req *requests.UpdateCurrencyRequest) (*responses.CurrencyResponse, error) {
	c, err := s.currencyRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("currency not found")
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	if req.Symbol != "" {
		c.Symbol = req.Symbol
	}
	if req.IsEnabled != nil {
		c.IsEnabled = *req.IsEnabled
	}
	if req.MinAmountMinor != nil {
		c.MinAmountMinor = *req.MinAmountMinor
	}
	if err := s.currencyRepo.Update(ctx, c); err != nil {
		return nil, err
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := s.currencyRepo.SetDefault(ctx, code); err != nil {
			return nil, err
		}
		c.IsDefault = true
	}
	s.audit(ctx, nil, "CURRENCY_UPDATED", code)
	return toCurrencyResponse(c), nil
}

func (s *currencyService) SetDefault(ctx context.Context, code string) error {
	if err := s.currencyRepo.SetDefault(ctx, code); err != nil {
		return err
	}
	s.audit(ctx, nil, "CURRENCY_SET_DEFAULT", code)
	return nil
}

// -------- Currency Requests --------

func (s *currencyService) RequestCurrency(ctx context.Context, actorID uuid.UUID, req *requests.RequestCurrencyRequest) (*responses.CurrencyRequestResponse, error) {
	// If already enabled, reject politely.
	existing, err := s.currencyRepo.FindByCode(ctx, req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.IsEnabled {
		return nil, fmt.Errorf("currency '%s' is already enabled", req.CurrencyCode)
	}
	// Prevent duplicate pending requests from same user for same currency.
	dup, err := s.currencyReqRepo.FindExisting(ctx, actorID, req.CurrencyCode)
	if err != nil {
		return nil, err
	}
	if dup != nil {
		return nil, fmt.Errorf("you already have a pending request for %s", req.CurrencyCode)
	}

	cr := &models.CurrencyRequest{
		RequestedBy:  actorID,
		CurrencyCode: req.CurrencyCode,
		Reason:       req.Reason,
		Status:       models.CurrencyRequestPending,
	}
	if err := s.currencyReqRepo.Create(ctx, cr); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "CURRENCY_REQUEST_CREATED", req.CurrencyCode)
	return toCurrencyRequestResponse(cr), nil
}

func (s *currencyService) ListMyRequests(ctx context.Context, actorID uuid.UUID, page, perPage int) ([]responses.CurrencyRequestResponse, int64, error) {
	list, total, err := s.currencyReqRepo.FindByRequester(ctx, actorID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.CurrencyRequestResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCurrencyRequestResponse(&list[i]))
	}
	return out, total, nil
}

func (s *currencyService) ListRequests(ctx context.Context, status string, page, perPage int) ([]responses.CurrencyRequestResponse, int64, error) {
	list, total, err := s.currencyReqRepo.FindAll(ctx, status, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.CurrencyRequestResponse, 0, len(list))
	for i := range list {
		r := toCurrencyRequestResponse(&list[i])
		r.RequesterName = list[i].Requester.FirstName + " " + list[i].Requester.LastName
		r.RequesterEmail = list[i].Requester.Email
		out = append(out, *r)
	}
	return out, total, nil
}

func (s *currencyService) ReviewRequest(ctx context.Context, actorID, requestID uuid.UUID, req *requests.ReviewCurrencyRequestRequest) (*responses.CurrencyRequestResponse, error) {
	cr, err := s.currencyReqRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if cr == nil {
		return nil, fmt.Errorf("currency request not found")
	}
	if cr.Status != models.CurrencyRequestPending {
		return nil, fmt.Errorf("currency request already reviewed")
	}
	cr.Status = models.CurrencyRequestStatus(req.Status)
	cr.ReviewedBy = &actorID
	cr.ReviewNote = req.ReviewNote

	if req.Status == "approved" {
		// Auto-create the currency as enabled if it doesn't exist.
		existing, _ := s.currencyRepo.FindByCode(ctx, cr.CurrencyCode)
		if existing == nil {
			newCur := &models.Currency{
				Code:      cr.CurrencyCode,
				Name:      cr.CurrencyCode, // admin can rename later
				IsEnabled: true,
			}
			if err := s.currencyRepo.Create(ctx, newCur); err != nil {
				return nil, err
			}
		} else if !existing.IsEnabled {
			existing.IsEnabled = true
			if err := s.currencyRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
		}
	}

	if err := s.currencyReqRepo.Update(ctx, cr); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "CURRENCY_REQUEST_REVIEWED", cr.CurrencyCode)
	return toCurrencyRequestResponse(cr), nil
}

func (s *currencyService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "currency",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log:", err)
	}
}

func toCurrencyResponse(c *models.Currency) *responses.CurrencyResponse {
	return &responses.CurrencyResponse{
		ID:             c.ID,
		Code:           c.Code,
		Name:           c.Name,
		Symbol:         c.Symbol,
		IsEnabled:      c.IsEnabled,
		IsDefault:      c.IsDefault,
		MinAmountMinor: c.MinAmountMinor,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func toCurrencyRequestResponse(r *models.CurrencyRequest) *responses.CurrencyRequestResponse {
	return &responses.CurrencyRequestResponse{
		ID:           r.ID,
		RequestedBy:  r.RequestedBy,
		CurrencyCode: r.CurrencyCode,
		Reason:       r.Reason,
		Status:       string(r.Status),
		ReviewedBy:   r.ReviewedBy,
		ReviewNote:   r.ReviewNote,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
