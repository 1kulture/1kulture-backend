package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	List(ctx context.Context) ([]models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
}
