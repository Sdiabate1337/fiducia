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

// ClientRepository handles database operations for clients
type ClientRepository struct {
	pool *pgxpool.Pool
}

// NewClientRepository creates a new repository
func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

// ClientFilter defines filtering options
type ClientFilter struct {
	CabinetID uuid.UUID
	Search    *string
	Phone     *string
	Limit     int
	Offset    int
}

// ClientList represents a paginated list result
type ClientList struct {
	Items   []models.Client `json:"items"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
}

// List returns clients with filtering and pagination
func (r *ClientRepository) List(ctx context.Context, filter ClientFilter) (*ClientList, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	baseQuery := `
		SELECT id, cabinet_id, name, siren, siret, phone, email, 
			   contact_name, address, notes, whatsapp_opted_in, 
			   whatsapp_opted_in_at, created_at, updated_at
		FROM clients
		WHERE cabinet_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM clients WHERE cabinet_id = $1`

	args := []any{filter.CabinetID}
	argPos := 2
	var conditions string

	if filter.Search != nil && *filter.Search != "" {
		conditions += fmt.Sprintf(" AND (name ILIKE $%d OR contact_name ILIKE $%d)", argPos, argPos)
		args = append(args, "%"+*filter.Search+"%")
		argPos++
	}

	if filter.Phone != nil && *filter.Phone != "" {
		conditions += fmt.Sprintf(" AND phone = $%d", argPos)
		args = append(args, *filter.Phone)
		argPos++
	}

	// Count total
	var total int
	err := r.pool.QueryRow(ctx, countQuery+conditions, args[:argPos-1]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count clients: %w", err)
	}

	// Query with pagination
	fullQuery := baseQuery + conditions +
		" ORDER BY name ASC" +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	items := make([]models.Client, 0)
	for rows.Next() {
		var c models.Client
		err := rows.Scan(
			&c.ID, &c.CabinetID, &c.Name, &c.SIREN, &c.SIRET,
			&c.Phone, &c.Email, &c.ContactName, &c.Address, &c.Notes,
			&c.WhatsAppOptedIn, &c.WhatsAppOptedInAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		items = append(items, c)
	}

	return &ClientList{
		Items:   items,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: filter.Offset+len(items) < total,
	}, nil
}

// GetByID returns a single client by ID
func (r *ClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	query := `
		SELECT id, cabinet_id, name, siren, siret, phone, email,
			   contact_name, address, notes, whatsapp_opted_in,
			   whatsapp_opted_in_at, created_at, updated_at
		FROM clients
		WHERE id = $1
	`

	var c models.Client
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.CabinetID, &c.Name, &c.SIREN, &c.SIRET,
		&c.Phone, &c.Email, &c.ContactName, &c.Address, &c.Notes,
		&c.WhatsAppOptedIn, &c.WhatsAppOptedInAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &c, nil
}

// GetByPhone returns a client by phone number
func (r *ClientRepository) GetByPhone(ctx context.Context, cabinetID uuid.UUID, phone string) (*models.Client, error) {
	query := `
		SELECT id, cabinet_id, name, siren, siret, phone, email,
			   contact_name, address, notes, whatsapp_opted_in,
			   whatsapp_opted_in_at, created_at, updated_at
		FROM clients
		WHERE cabinet_id = $1 AND phone = $2
	`

	var c models.Client
	err := r.pool.QueryRow(ctx, query, cabinetID, phone).Scan(
		&c.ID, &c.CabinetID, &c.Name, &c.SIREN, &c.SIRET,
		&c.Phone, &c.Email, &c.ContactName, &c.Address, &c.Notes,
		&c.WhatsAppOptedIn, &c.WhatsAppOptedInAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client by phone: %w", err)
	}

	return &c, nil
}

// Create inserts a new client
func (r *ClientRepository) Create(ctx context.Context, c *models.Client) error {
	query := `
		INSERT INTO clients (
			id, cabinet_id, name, siren, siret, phone, email,
			contact_name, address, notes, whatsapp_opted_in,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		c.ID, c.CabinetID, c.Name, c.SIREN, c.SIRET,
		c.Phone, c.Email, c.ContactName, c.Address, c.Notes,
		c.WhatsAppOptedIn, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// Update updates an existing client
func (r *ClientRepository) Update(ctx context.Context, c *models.Client) error {
	query := `
		UPDATE clients SET
			name = $2, siren = $3, siret = $4, phone = $5, email = $6,
			contact_name = $7, address = $8, notes = $9, whatsapp_opted_in = $10,
			whatsapp_opted_in_at = $11, updated_at = $12
		WHERE id = $1
	`

	c.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		c.ID, c.Name, c.SIREN, c.SIRET, c.Phone, c.Email,
		c.ContactName, c.Address, c.Notes, c.WhatsAppOptedIn,
		c.WhatsAppOptedInAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// Delete removes a client
func (r *ClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM clients WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// SetWhatsAppOptIn updates the WhatsApp opt-in status
func (r *ClientRepository) SetWhatsAppOptIn(ctx context.Context, id uuid.UUID, optedIn bool) error {
	var optedInAt *time.Time
	if optedIn {
		now := time.Now()
		optedInAt = &now
	}

	query := `
		UPDATE clients SET 
			whatsapp_opted_in = $2, 
			whatsapp_opted_in_at = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, optedIn, optedInAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update opt-in status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// FindOrCreateByName finds a client by name or creates a new one
func (r *ClientRepository) FindOrCreateByName(ctx context.Context, cabinetID uuid.UUID, name string) (*models.Client, bool, error) {
	// Try to find existing
	query := `
		SELECT id, cabinet_id, name, siren, siret, phone, email,
			   contact_name, address, notes, whatsapp_opted_in,
			   whatsapp_opted_in_at, created_at, updated_at
		FROM clients
		WHERE cabinet_id = $1 AND LOWER(name) = LOWER($2)
		LIMIT 1
	`

	var c models.Client
	err := r.pool.QueryRow(ctx, query, cabinetID, name).Scan(
		&c.ID, &c.CabinetID, &c.Name, &c.SIREN, &c.SIRET,
		&c.Phone, &c.Email, &c.ContactName, &c.Address, &c.Notes,
		&c.WhatsAppOptedIn, &c.WhatsAppOptedInAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == nil {
		return &c, false, nil // Found existing
	}
	if err != pgx.ErrNoRows {
		return nil, false, fmt.Errorf("failed to find client: %w", err)
	}

	// Create new client
	c = models.Client{
		ID:        uuid.New(),
		CabinetID: cabinetID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = r.Create(ctx, &c)
	if err != nil {
		return nil, false, err
	}

	return &c, true, nil // Created new
}
