package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fiducia/backend/internal/database"
	"github.com/fiducia/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CabinetRepository struct {
	db *database.DB
}

func NewCabinetRepository(db *database.DB) *CabinetRepository {
	return &CabinetRepository{db: db}
}

func (r *CabinetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Cabinet, error) {
	query := `
		SELECT id, name, siret, email, phone, address, settings, onboarding_completed, created_at, updated_at
		FROM cabinets
		WHERE id = $1
	`
	var c models.Cabinet
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.SIRET, &c.Email, &c.Phone, &c.Address, &c.Settings, &c.OnboardingCompleted, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cabinet: %w", err)
	}
	return &c, nil
}

func (r *CabinetRepository) Update(ctx context.Context, cabinet *models.Cabinet) error {
	query := `
		UPDATE cabinets
		SET name = $1, siret = $2, email = $3, phone = $4, address = $5, settings = $6, onboarding_completed = $7, updated_at = $8
		WHERE id = $9
	`
	_, err := r.db.Pool.Exec(ctx, query,
		cabinet.Name,
		cabinet.SIRET,
		cabinet.Email,
		cabinet.Phone,
		cabinet.Address,
		cabinet.Settings,
		cabinet.OnboardingCompleted,
		time.Now(),
		cabinet.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update cabinet: %w", err)
	}
	return nil
}

// Add Create if needed, but for now AuthSvc handles creation
