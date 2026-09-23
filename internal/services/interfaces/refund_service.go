package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type RefundService interface {
	// RequestRefund is invoked by a buyer (or an organizer) to request a refund for
	// specific tickets in an order.
	RequestRefund(ctx context.Context, actorID, orderID uuid.UUID, req *requests.RefundRequestPayload) (*responses.RefundResponse, error)

	// DecideRefund is invoked by an event organizer / co-organizer / admin to approve or reject a refund.
	DecideRefund(ctx context.Context, actorID, refundID uuid.UUID, req *requests.RefundDecisionRequest) (*responses.RefundResponse, error)

	// GetRefund returns a refund record. Buyer, organizer of the event, and admins can view.
	GetRefund(ctx context.Context, viewerID, refundID uuid.UUID) (*responses.RefundResponse, error)

	// ListByOrder returns all refunds for an order.
	ListByOrder(ctx context.Context, viewerID, orderID uuid.UUID) ([]responses.RefundResponse, error)

	// ListByStatus admin-only listing of refunds.
	ListByStatus(ctx context.Context, status string, page, perPage int) ([]responses.RefundResponse, int64, error)
}
