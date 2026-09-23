package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type CheckInService interface {
	// ScanByQR is used by staff with the scanner app; body contains the QR payload.
	ScanByQR(ctx context.Context, actorID, eventID uuid.UUID, req *requests.CheckInRequest, ip string) (*responses.CheckInResponse, error)

	// ManualCheckIn is used when the QR cannot be scanned (e.g. broken screen).
	ManualCheckIn(ctx context.Context, actorID, eventID uuid.UUID, req *requests.ManualCheckInRequest, ip string) (*responses.CheckInResponse, error)

	// ListCheckIns lists check-ins for an event (organizer/staff only).
	ListCheckIns(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.CheckInRecordResponse, int64, error)

	// Stats returns aggregated check-in stats.
	Stats(ctx context.Context, actorID, eventID uuid.UUID) (*responses.CheckInStatsResponse, error)
}
