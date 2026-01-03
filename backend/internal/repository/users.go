package repository

import (
	"context"
	"fmt"

	"github.com/fiducia/backend/internal/database"
	"github.com/fiducia/backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, cabinet_id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.CabinetID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT u.id, u.cabinet_id, u.email, u.password_hash, u.full_name, u.role, u.created_at, u.updated_at, c.onboarding_completed
		FROM users u
		JOIN cabinets c ON u.cabinet_id = c.id
		WHERE u.email = $1
	`
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.CabinetID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.OnboardingCompleted,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.cabinet_id, u.email, u.password_hash, u.full_name, u.role, u.created_at, u.updated_at, c.onboarding_completed
		FROM users u
		JOIN cabinets c ON u.cabinet_id = c.id
		WHERE u.id = $1
	`
	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.CabinetID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.OnboardingCompleted,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
