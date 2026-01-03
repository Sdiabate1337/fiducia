package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/models"
)

// PendingLineRepository handles database operations for pending lines
type PendingLineRepository struct {
	pool *pgxpool.Pool
}

// NewPendingLineRepository creates a new repository
func NewPendingLineRepository(pool *pgxpool.Pool) *PendingLineRepository {
	return &PendingLineRepository{pool: pool}
}

// PendingLineFilter defines filtering options
type PendingLineFilter struct {
	CabinetID  uuid.UUID
	ClientID   *uuid.UUID
	Status     *models.PendingLineStatus
	DateFrom   *time.Time
	DateTo     *time.Time
	AmountMin  *float64
	AmountMax  *float64
	Search     *string
	AssignedTo *uuid.UUID
	Limit      int
	Offset     int
}

// PendingLineList represents a paginated list result
type PendingLineList struct {
	Items   []models.PendingLine `json:"items"`
	Total   int                  `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
	HasMore bool                 `json:"has_more"`
}

// List returns pending lines with filtering and pagination
func (r *PendingLineRepository) List(ctx context.Context, filter PendingLineFilter) (*PendingLineList, error) {
	// Set defaults
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	// Build query
	baseQuery := `
		SELECT 
			pl.id, pl.cabinet_id, pl.client_id, pl.amount, pl.transaction_date,
			pl.bank_label, pl.account_number, pl.import_batch_id, pl.source_file,
			pl.source_row_number, pl.status, pl.last_contacted_at, pl.contact_count,
			pl.assigned_to, pl.created_at, pl.updated_at,
			c.id as client_id, c.name as client_name, c.phone as client_phone,
			ce.status as campaign_status, ce.next_step_scheduled_at, ce.current_step_order
		FROM pending_lines pl
		LEFT JOIN clients c ON pl.client_id = c.id
		LEFT JOIN campaign_executions ce ON pl.id = ce.pending_line_id AND ce.status IN ('pending', 'running', 'stopped', 'completed')
		WHERE pl.cabinet_id = $1
	`

	countQuery := `SELECT COUNT(*) FROM pending_lines pl WHERE pl.cabinet_id = $1`

	args := []any{filter.CabinetID}
	argPos := 2

	// Add filters
	var conditions string

	if filter.ClientID != nil {
		conditions += fmt.Sprintf(" AND pl.client_id = $%d", argPos)
		args = append(args, *filter.ClientID)
		argPos++
	}

	if filter.Status != nil {
		conditions += fmt.Sprintf(" AND pl.status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.DateFrom != nil {
		conditions += fmt.Sprintf(" AND pl.transaction_date >= $%d", argPos)
		args = append(args, *filter.DateFrom)
		argPos++
	}

	if filter.DateTo != nil {
		conditions += fmt.Sprintf(" AND pl.transaction_date <= $%d", argPos)
		args = append(args, *filter.DateTo)
		argPos++
	}

	if filter.AmountMin != nil {
		conditions += fmt.Sprintf(" AND pl.amount >= $%d", argPos)
		args = append(args, *filter.AmountMin)
		argPos++
	}

	if filter.AmountMax != nil {
		conditions += fmt.Sprintf(" AND pl.amount <= $%d", argPos)
		args = append(args, *filter.AmountMax)
		argPos++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions += fmt.Sprintf(" AND pl.bank_label ILIKE $%d", argPos)
		args = append(args, "%"+*filter.Search+"%")
		argPos++
	}

	if filter.AssignedTo != nil {
		conditions += fmt.Sprintf(" AND pl.assigned_to = $%d", argPos)
		args = append(args, *filter.AssignedTo)
		argPos++
	}

	// Get total count
	var total int
	err := r.pool.QueryRow(ctx, countQuery+conditions, args[:argPos-1]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending lines: %w", err)
	}

	// Add pagination and ordering
	fullQuery := baseQuery + conditions +
		" ORDER BY pl.transaction_date DESC, pl.created_at DESC" +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	// Execute query
	// DEBUG: Log query to file
	debugMsg := fmt.Sprintf("Query: %s\nArgs: %v\n", fullQuery, args)
	_ = os.WriteFile("/tmp/query_debug.txt", []byte(debugMsg), 0644)

	rows, err := r.pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending lines: %w", err)
	}
	defer rows.Close()

	items := make([]models.PendingLine, 0)
	for rows.Next() {
		var pl models.PendingLine
		var clientID, clientName, clientPhone *string

		err := rows.Scan(
			&pl.ID, &pl.CabinetID, &pl.ClientID, &pl.Amount, &pl.TransactionDate,
			&pl.BankLabel, &pl.AccountNumber, &pl.ImportBatchID, &pl.SourceFile,
			&pl.SourceRowNumber, &pl.Status, &pl.LastContactedAt, &pl.ContactCount,
			&pl.AssignedTo, &pl.CreatedAt, &pl.UpdatedAt,
			&clientID, &clientName, &clientPhone,
			&pl.CampaignStatus, &pl.NextStepScheduledAt, &pl.CampaignCurrentStep,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending line: %w", err)
		}

		// Attach client if present
		if clientID != nil && clientName != nil {
			clientUUID, _ := uuid.Parse(*clientID)
			pl.Client = &models.Client{
				ID:    clientUUID,
				Name:  *clientName,
				Phone: clientPhone,
			}
		}

		items = append(items, pl)
	}

	return &PendingLineList{
		Items:   items,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: filter.Offset+len(items) < total,
	}, nil
}

// GetByID returns a single pending line by ID
func (r *PendingLineRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PendingLine, error) {
	query := `
		SELECT 
			pl.id, pl.cabinet_id, pl.client_id, pl.amount, pl.transaction_date,
			pl.bank_label, pl.account_number, pl.import_batch_id, pl.source_file,
			pl.source_row_number, pl.status, pl.last_contacted_at, pl.contact_count,
			pl.assigned_to, pl.created_at, pl.updated_at,
			c.id as client_id, c.name as client_name, c.phone as client_phone
		FROM pending_lines pl
		LEFT JOIN clients c ON pl.client_id = c.id
		WHERE pl.id = $1
	`

	var pl models.PendingLine
	var clientID, clientName, clientPhone *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pl.ID, &pl.CabinetID, &pl.ClientID, &pl.Amount, &pl.TransactionDate,
		&pl.BankLabel, &pl.AccountNumber, &pl.ImportBatchID, &pl.SourceFile,
		&pl.SourceRowNumber, &pl.Status, &pl.LastContactedAt, &pl.ContactCount,
		&pl.AssignedTo, &pl.CreatedAt, &pl.UpdatedAt,
		&clientID, &clientName, &clientPhone,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pending line: %w", err)
	}

	// Attach client if present
	if clientID != nil && clientName != nil {
		clientUUID, _ := uuid.Parse(*clientID)
		pl.Client = &models.Client{
			ID:    clientUUID,
			Name:  *clientName,
			Phone: clientPhone,
		}
	}

	return &pl, nil
}

// Create inserts a new pending line
func (r *PendingLineRepository) Create(ctx context.Context, pl *models.PendingLine) error {
	query := `
		INSERT INTO pending_lines (
			id, cabinet_id, client_id, amount, transaction_date, bank_label,
			account_number, import_batch_id, source_file, source_row_number,
			status, contact_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	pl.CreatedAt = time.Now()
	pl.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		pl.ID, pl.CabinetID, pl.ClientID, pl.Amount, pl.TransactionDate,
		pl.BankLabel, pl.AccountNumber, pl.ImportBatchID, pl.SourceFile,
		pl.SourceRowNumber, pl.Status, pl.ContactCount, pl.CreatedAt, pl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create pending line: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple pending lines in a transaction
func (r *PendingLineRepository) CreateBatch(ctx context.Context, lines []models.PendingLine) error {
	if len(lines) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO pending_lines (
			id, cabinet_id, client_id, amount, transaction_date, bank_label,
			account_number, import_batch_id, source_file, source_row_number,
			status, contact_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	now := time.Now()
	for i := range lines {
		if lines[i].ID == uuid.Nil {
			lines[i].ID = uuid.New()
		}
		lines[i].CreatedAt = now
		lines[i].UpdatedAt = now

		_, err := tx.Exec(ctx, query,
			lines[i].ID, lines[i].CabinetID, lines[i].ClientID, lines[i].Amount,
			lines[i].TransactionDate, lines[i].BankLabel, lines[i].AccountNumber,
			lines[i].ImportBatchID, lines[i].SourceFile, lines[i].SourceRowNumber,
			lines[i].Status, lines[i].ContactCount, lines[i].CreatedAt, lines[i].UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert line %d: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Update updates an existing pending line
func (r *PendingLineRepository) Update(ctx context.Context, pl *models.PendingLine) error {
	query := `
		UPDATE pending_lines SET
			client_id = $2,
			status = $3,
			last_contacted_at = $4,
			contact_count = $5,
			assigned_to = $6,
			updated_at = $7
		WHERE id = $1
	`

	pl.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		pl.ID, pl.ClientID, pl.Status, pl.LastContactedAt,
		pl.ContactCount, pl.AssignedTo, pl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update pending line: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending line not found")
	}

	return nil
}

// UpdateStatus updates only the status of a pending line
func (r *PendingLineRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.PendingLineStatus) error {
	query := `UPDATE pending_lines SET status = $2, updated_at = $3 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending line not found")
	}

	return nil
}

// Delete removes a pending line
func (r *PendingLineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM pending_lines WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete pending line: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending line not found")
	}

	return nil
}

// GetStats returns statistics for a cabinet's pending lines
func (r *PendingLineRepository) GetStats(ctx context.Context, cabinetID uuid.UUID) (map[string]any, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'contacted') as contacted,
			COUNT(*) FILTER (WHERE status = 'received') as received,
			COUNT(*) FILTER (WHERE status = 'validated') as validated,
			COUNT(*) FILTER (WHERE status = 'rejected') as rejected,
			COALESCE(SUM(amount) FILTER (WHERE status = 'pending'), 0) as pending_amount,
			COALESCE(SUM(amount) FILTER (WHERE status = 'validated'), 0) as validated_amount
		FROM pending_lines
		WHERE cabinet_id = $1
	`

	var total, pending, contacted, received, validated, rejected int
	var pendingAmount, validatedAmount float64

	err := r.pool.QueryRow(ctx, query, cabinetID).Scan(
		&total, &pending, &contacted, &received, &validated, &rejected,
		&pendingAmount, &validatedAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return map[string]any{
		"total":            total,
		"pending":          pending,
		"contacted":        contacted,
		"received":         received,
		"validated":        validated,
		"rejected":         rejected,
		"pending_amount":   pendingAmount,
		"validated_amount": validatedAmount,
	}, nil
}

// ListByClient returns all pending lines for a specific client
func (r *PendingLineRepository) ListByClient(ctx context.Context, clientID uuid.UUID) ([]*models.PendingLine, error) {
	query := `
		SELECT 
			id, cabinet_id, client_id, amount, transaction_date,
			bank_label, account_number, import_batch_id, source_file,
			source_row_number, status, last_contacted_at, contact_count,
			assigned_to, created_at, updated_at
		FROM pending_lines
		WHERE client_id = $1
		ORDER BY transaction_date DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list by client: %w", err)
	}
	defer rows.Close()

	var lines []*models.PendingLine
	for rows.Next() {
		var pl models.PendingLine
		if err := rows.Scan(
			&pl.ID, &pl.CabinetID, &pl.ClientID, &pl.Amount, &pl.TransactionDate,
			&pl.BankLabel, &pl.AccountNumber, &pl.ImportBatchID, &pl.SourceFile,
			&pl.SourceRowNumber, &pl.Status, &pl.LastContactedAt, &pl.ContactCount,
			&pl.AssignedTo, &pl.CreatedAt, &pl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		lines = append(lines, &pl)
	}

	return lines, nil
}
