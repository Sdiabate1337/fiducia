package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/models"
)

type CampaignRepository struct {
	pool *pgxpool.Pool
}

func NewCampaignRepository(pool *pgxpool.Pool) *CampaignRepository {
	return &CampaignRepository{pool: pool}
}

// Create inserts a new campaign and its steps
func (r *CampaignRepository) Create(ctx context.Context, c *models.Campaign) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	// Insert Campaign
	_, err = tx.Exec(ctx, `
        INSERT INTO campaigns (id, cabinet_id, name, trigger_type, is_active, quiet_hours_enabled, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, c.ID, c.CabinetID, c.Name, c.TriggerType, c.IsActive, c.QuietHoursEnabled, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert campaign: %w", err)
	}

	// Insert Steps
	for i := range c.Steps {
		s := &c.Steps[i]
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		s.CampaignID = c.ID
		s.CreatedAt = time.Now()

		_, err = tx.Exec(ctx, `
            INSERT INTO campaign_steps (id, campaign_id, step_order, delay_hours, channel, template_id, config, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, s.ID, s.CampaignID, s.StepOrder, s.DelayHours, s.Channel, s.TemplateID, s.Config, s.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert step %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID returns a campaign with its steps
func (r *CampaignRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Campaign, error) {
	var c models.Campaign
	err := r.pool.QueryRow(ctx, `
        SELECT id, cabinet_id, name, trigger_type, is_active, quiet_hours_enabled, created_at, updated_at
        FROM campaigns WHERE id = $1
    `, id).Scan(&c.ID, &c.CabinetID, &c.Name, &c.TriggerType, &c.IsActive, &c.QuietHoursEnabled, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	// Get Steps
	rows, err := r.pool.Query(ctx, `
        SELECT id, campaign_id, step_order, delay_hours, channel, template_id, config, created_at
        FROM campaign_steps WHERE campaign_id = $1 ORDER BY step_order ASC
    `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c.Steps = make([]models.CampaignStep, 0)
	for rows.Next() {
		var s models.CampaignStep
		if err := rows.Scan(&s.ID, &s.CampaignID, &s.StepOrder, &s.DelayHours, &s.Channel, &s.TemplateID, &s.Config, &s.CreatedAt); err != nil {
			return nil, err
		}
		c.Steps = append(c.Steps, s)
	}

	return &c, nil
}

// List returns all campaigns for a cabinet
func (r *CampaignRepository) List(ctx context.Context, cabinetID uuid.UUID) ([]models.Campaign, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, cabinet_id, name, trigger_type, is_active, quiet_hours_enabled, created_at, updated_at
        FROM campaigns WHERE cabinet_id = $1 ORDER BY created_at DESC
    `, cabinetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.CabinetID, &c.Name, &c.TriggerType, &c.IsActive, &c.QuietHoursEnabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// Update updates a campaign details and performs a full replace of its steps if provided
func (r *CampaignRepository) Update(ctx context.Context, c *models.Campaign) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	c.UpdatedAt = time.Now()
	res, err := tx.Exec(ctx, `
        UPDATE campaigns SET name=$2, trigger_type=$3, is_active=$4, quiet_hours_enabled=$5, updated_at=$6
        WHERE id=$1
    `, c.ID, c.Name, c.TriggerType, c.IsActive, c.QuietHoursEnabled, c.UpdatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	// If steps slice is not nil, we assume we want to sync steps (delete all and re-insert)
	if c.Steps != nil {
		// Delete existing
		_, err = tx.Exec(ctx, `DELETE FROM campaign_steps WHERE campaign_id = $1`, c.ID)
		if err != nil {
			return err
		}
		// Insert new
		for _, s := range c.Steps {
			if s.ID == uuid.Nil {
				s.ID = uuid.New()
			}
			s.CampaignID = c.ID
			_, err = tx.Exec(ctx, `
                INSERT INTO campaign_steps (id, campaign_id, step_order, delay_hours, channel, template_id, config, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            `, s.ID, s.CampaignID, s.StepOrder, s.DelayHours, s.Channel, s.TemplateID, s.Config, time.Now())
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a campaign and cascade steps
func (r *CampaignRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, "DELETE FROM campaigns WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}
	return nil
}
