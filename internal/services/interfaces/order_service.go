package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type OrderService interface {
	// CreateOrder validates items, applies promo, resolves commission,
	// atomically reserves stock, and creates a pending Order.
	CreateOrder(ctx context.Context, userID uuid.UUID, req *requests.CreateOrderRequest) (*responses.OrderResponse, error)

	// InitializePayment creates a provider transaction for a pending order.
	InitializePayment(ctx context.Context, userID, orderID uuid.UUID, req *requests.InitializePaymentRequest) (*responses.InitializePaymentResponse, error)

	// HandleWebhook processes a provider callback (paystack currently).
	// rawBody must be the exact bytes; signature comes from the provider header.
	HandleWebhook(ctx context.Context, provider string, rawBody []byte, signature string) error

	// GetOrder returns a single order (owner or event organizer only).
	GetOrder(ctx context.Context, viewerID, orderID uuid.UUID) (*responses.OrderResponse, error)

	// ListMyOrders returns the buyer's orders.
	ListMyOrders(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.OrderResponse, int64, error)

	// ListEventOrders returns orders for an event (organizer only).
	ListEventOrders(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.OrderResponse, int64, error)

	// ExpirePendingOrders releases reservations for orders past their hold window.
	// Returns number of orders expired.
	ExpirePendingOrders(ctx context.Context, limit int) (int, error)
}
