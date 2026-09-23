package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type TicketTransferService interface {
	// InitiateTransfer is called by the current ticket owner to send the ticket to another email.
	InitiateTransfer(ctx context.Context, actorID, ticketID uuid.UUID, req *requests.InitiateTransferRequest) (*responses.TicketTransferResponse, error)

	// AcceptTransfer is called by the recipient (must be logged in with the same email).
	AcceptTransfer(ctx context.Context, actorID uuid.UUID, req *requests.AcceptTransferRequest) (*responses.TicketResponse, error)

	// DeclineTransfer is called by the recipient to reject the transfer.
	DeclineTransfer(ctx context.Context, actorID uuid.UUID, req *requests.DeclineTransferRequest) (*responses.TicketTransferResponse, error)

	// CancelTransfer is called by the sender to cancel a pending transfer.
	CancelTransfer(ctx context.Context, actorID, transferID uuid.UUID) (*responses.TicketTransferResponse, error)

	// GetTransfer returns a transfer by ID (sender or recipient).
	GetTransfer(ctx context.Context, viewerID, transferID uuid.UUID) (*responses.TicketTransferResponse, error)

	// GetTransferByToken is a public lookup for the recipient to see the transfer details.
	GetTransferByToken(ctx context.Context, token string) (*responses.TicketTransferResponse, error)

	// ListMyTransfers lists transfers initiated by the current user.
	ListMyTransfers(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.TicketTransferResponse, int64, error)
}
