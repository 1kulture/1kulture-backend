package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type CurrencyRepository interface {
	Create(ctx context.Context, currency *models.Currency) error
	FindByCode(ctx context.Context, code string) (*models.Currency, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Currency, error)
	FindAll(ctx context.Context, onlyEnabled bool) ([]models.Currency, error)
	FindDefault(ctx context.Context) (*models.Currency, error)
	Update(ctx context.Context, currency *models.Currency) error
	SetDefault(ctx context.Context, code string) error
}

type CurrencyRequestRepository interface {
	Create(ctx context.Context, req *models.CurrencyRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.CurrencyRequest, error)
	FindByRequester(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.CurrencyRequest, int64, error)
	FindAll(ctx context.Context, status string, page, perPage int) ([]models.CurrencyRequest, int64, error)
	FindExisting(ctx context.Context, userID uuid.UUID, code string) (*models.CurrencyRequest, error)
	Update(ctx context.Context, req *models.CurrencyRequest) error
}
