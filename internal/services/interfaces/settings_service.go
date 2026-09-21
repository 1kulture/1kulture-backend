package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type SettingService interface {
	Create(ctx context.Context, actorID uuid.UUID, req *requests.CreateSettingRequest) (*responses.SettingResponse, error)
	GetByKey(ctx context.Context, key string) (*responses.SettingResponse, error)
	List(ctx context.Context, category string) ([]responses.SettingResponse, error)
	ListPublic(ctx context.Context) (*responses.PublicSettingsResponse, error)
	Update(ctx context.Context, actorID uuid.UUID, req *requests.UpdateSettingRequest) (*responses.SettingResponse, error)
	BulkUpdate(ctx context.Context, actorID uuid.UUID, req *requests.BulkUpdateSettingsRequest) ([]responses.SettingResponse, error)
	Delete(ctx context.Context, key string) error

	// Typed accessors used by other services
	GetInt(ctx context.Context, key string, fallback int) int
	GetString(ctx context.Context, key string, fallback string) string
	GetBool(ctx context.Context, key string, fallback bool) bool
}
