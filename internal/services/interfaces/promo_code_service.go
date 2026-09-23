package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type PromoCodeService interface {
	Create(ctx context.Context, actorID, eventID uuid.UUID, req *requests.PromoCodeCreateRequest) (*responses.PromoCodeResponse, error)
	List(ctx context.Context, actorID, eventID uuid.UUID) ([]responses.PromoCodeResponse, error)
	Update(ctx context.Context, actorID, eventID, promoID uuid.UUID, req *requests.PromoCodeUpdateRequest) (*responses.PromoCodeResponse, error)
	Delete(ctx context.Context, actorID, eventID, promoID uuid.UUID) error

	// Validate checks a promo code against a subtotal without consuming it.
	Validate(ctx context.Context, userID uuid.UUID, req *requests.ValidatePromoRequest) (*responses.PromoValidationResponse, error)
}
