package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type CurrencyService interface {
	Create(ctx context.Context, actorID uuid.UUID, req *requests.CreateCurrencyRequest) (*responses.CurrencyResponse, error)
	GetByCode(ctx context.Context, code string) (*responses.CurrencyResponse, error)
	List(ctx context.Context, onlyEnabled bool) ([]responses.CurrencyResponse, error)
	Update(ctx context.Context, code string, req *requests.UpdateCurrencyRequest) (*responses.CurrencyResponse, error)
	SetDefault(ctx context.Context, code string) error

	// Currency requests
	RequestCurrency(ctx context.Context, actorID uuid.UUID, req *requests.RequestCurrencyRequest) (*responses.CurrencyRequestResponse, error)
	ListMyRequests(ctx context.Context, actorID uuid.UUID, page, perPage int) ([]responses.CurrencyRequestResponse, int64, error)
	ListRequests(ctx context.Context, status string, page, perPage int) ([]responses.CurrencyRequestResponse, int64, error)
	ReviewRequest(ctx context.Context, actorID, requestID uuid.UUID, req *requests.ReviewCurrencyRequestRequest) (*responses.CurrencyRequestResponse, error)
}
