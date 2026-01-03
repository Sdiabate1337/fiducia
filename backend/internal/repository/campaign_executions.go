package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/models"
)

type CampaignExecutionRepository struct {
	pool *pgxpool.Pool
}

func NewCampaignExecutionRepository(pool *pgxpool.Pool) *CampaignExecutionRepository {
	return &CampaignExecutionRepository{pool: pool}
}

// Create starts tracking a campaign for a line
func (r *CampaignExecutionRepository) Create(ctx context.Context, ex *models.CampaignExecution) error {
	if ex.ID == uuid.Nil {
		ex.ID = uuid.New()
	}
	ex.CreatedAt = time.Now()
	ex.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
        INSERT INTO campaign_executions (
            id, campaign_id, pending_line_id, current_step_order, 
            status, stop_reason, last_step_executed_at, next_step_scheduled_at, 
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, ex.ID, ex.CampaignID, ex.PendingLineID, ex.CurrentStepOrder,
		ex.Status, ex.StopReason, ex.LastStepExecutedAt, ex.NextStepScheduledAt,
		ex.CreatedAt, ex.UpdatedAt)

	return err
}

// Update updates the execution state
func (r *CampaignExecutionRepository) Update(ctx context.Context, ex *models.CampaignExecution) error {
	ex.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
        UPDATE campaign_executions SET 
            current_step_order=$2, status=$3, stop_reason=$4, 
            last_step_executed_at=$5, next_step_scheduled_at=$6, updated_at=$7
        WHERE id=$1
    `, ex.ID, ex.CurrentStepOrder, ex.Status, ex.StopReason,
		ex.LastStepExecutedAt, ex.NextStepScheduledAt, ex.UpdatedAt)
	return err
}

// FindActive returns all executions that are running or pending
func (r *CampaignExecutionRepository) FindActive(ctx context.Context) ([]models.CampaignExecution, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, campaign_id, pending_line_id, current_step_order, 
               status, stop_reason, last_step_executed_at, next_step_scheduled_at, 
               created_at, updated_at
        FROM campaign_executions 
        WHERE status IN ('pending', 'running')
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.CampaignExecution
	for rows.Next() {
		var ex models.CampaignExecution
		err := rows.Scan(
			&ex.ID, &ex.CampaignID, &ex.PendingLineID, &ex.CurrentStepOrder,
			&ex.Status, &ex.StopReason, &ex.LastStepExecutedAt, &ex.NextStepScheduledAt,
			&ex.CreatedAt, &ex.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, ex)
	}
	return list, nil
}

// FindUnenrolledLines finds pending lines that match the trigger but are NOT yet in campaign_executions
// Simplified for MVP: finds all 'pending' lines not in executions table for this campaign
func (r *CampaignExecutionRepository) FindUnenrolledLines(ctx context.Context, campaignID uuid.UUID, cabinetID uuid.UUID) ([]uuid.UUID, error) {
	// join with pending_lines to filter by cabinet_id and status=pending
	rows, err := r.pool.Query(ctx, `
        SELECT pl.id
        FROM pending_lines pl
        LEFT JOIN campaign_executions ce ON pl.id = ce.pending_line_id AND ce.campaign_id = $1
        WHERE pl.cabinet_id = $2 
          AND pl.status = 'pending' 
          AND ce.id IS NULL
    `, campaignID, cabinetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
