package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
)

type WaitlistService interface {
	AddToWaitlist(ctx context.Context, req *requests.WaitlistRequest) error
}
