package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, page, perPage int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Preload("Roles").Limit(perPage).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) UpdateEmailVerification(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("email_verified_at", verifiedAt).Error; err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLoginAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", lastLoginAt).Error; err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

func (r *userRepository) WithRoles(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user with roles: %w", err)
	}
	return &user, nil
}
