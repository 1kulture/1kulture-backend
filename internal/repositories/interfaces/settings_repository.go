package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
)

type SettingRepository interface {
	Create(ctx context.Context, setting *models.Setting) error
	FindByKey(ctx context.Context, key string) (*models.Setting, error)
	FindAll(ctx context.Context, category string) ([]models.Setting, error)
	FindPublic(ctx context.Context) ([]models.Setting, error)
	Update(ctx context.Context, setting *models.Setting) error
	Upsert(ctx context.Context, setting *models.Setting) error
	Delete(ctx context.Context, key string) error
}
