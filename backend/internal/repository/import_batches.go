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

// ImportBatchRepository handles database operations for import batches
type ImportBatchRepository struct {
	pool *pgxpool.Pool
}

// NewImportBatchRepository creates a new repository
func NewImportBatchRepository(pool *pgxpool.Pool) *ImportBatchRepository {
	return &ImportBatchRepository{pool: pool}
}

// Create inserts a new import batch
func (r *ImportBatchRepository) Create(ctx context.Context, batch *models.ImportBatch) error {
	query := `
		INSERT INTO import_batches (
			id, cabinet_id, imported_by, filename, file_type,
			total_rows, imported_rows, failed_rows, errors, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	if batch.ID == uuid.Nil {
		batch.ID = uuid.New()
	}
	batch.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		batch.ID, batch.CabinetID, batch.ImportedBy, batch.Filename,
		batch.FileType, batch.TotalRows, batch.ImportedRows, batch.FailedRows,
		batch.Errors, batch.Status, batch.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create import batch: %w", err)
	}

	return nil
}

// GetByID returns a single import batch by ID
func (r *ImportBatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ImportBatch, error) {
	query := `
		SELECT id, cabinet_id, imported_by, filename, file_type,
			   total_rows, imported_rows, failed_rows, errors, status,
			   created_at, completed_at
		FROM import_batches
		WHERE id = $1
	`

	var b models.ImportBatch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.CabinetID, &b.ImportedBy, &b.Filename, &b.FileType,
		&b.TotalRows, &b.ImportedRows, &b.FailedRows, &b.Errors, &b.Status,
		&b.CreatedAt, &b.CompletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get import batch: %w", err)
	}

	return &b, nil
}

// UpdateStatus updates the status and results of an import batch
func (r *ImportBatchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, importedRows, failedRows int, errors map[string]any) error {
	query := `
		UPDATE import_batches SET
			status = $2, imported_rows = $3, failed_rows = $4,
			errors = $5, completed_at = $6
		WHERE id = $1
	`

	var completedAt *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now()
		completedAt = &now
	}

	result, err := r.pool.Exec(ctx, query, id, status, importedRows, failedRows, errors, completedAt)
	if err != nil {
		return fmt.Errorf("failed to update import batch: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("import batch not found")
	}

	return nil
}

// List returns recent import batches for a cabinet
func (r *ImportBatchRepository) List(ctx context.Context, cabinetID uuid.UUID, limit int) ([]models.ImportBatch, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, cabinet_id, imported_by, filename, file_type,
			   total_rows, imported_rows, failed_rows, errors, status,
			   created_at, completed_at
		FROM import_batches
		WHERE cabinet_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, cabinetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query import batches: %w", err)
	}
	defer rows.Close()

	var batches []models.ImportBatch
	for rows.Next() {
		var b models.ImportBatch
		err := rows.Scan(
			&b.ID, &b.CabinetID, &b.ImportedBy, &b.Filename, &b.FileType,
			&b.TotalRows, &b.ImportedRows, &b.FailedRows, &b.Errors, &b.Status,
			&b.CreatedAt, &b.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import batch: %w", err)
		}
		batches = append(batches, b)
	}

	return batches, nil
}
