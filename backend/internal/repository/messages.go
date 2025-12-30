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

// MessageRepository handles database operations for messages
type MessageRepository struct {
	pool *pgxpool.Pool
}

// NewMessageRepository creates a new repository
func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

// Create inserts a new message
func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (
			id, pending_line_id, client_id, direction, message_type,
			content, media_url, template_name, template_params,
			wa_message_id, status, scheduled_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	msg.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.PendingLineID, msg.ClientID, msg.Direction, msg.MessageType,
		msg.Content, msg.MediaURL, msg.TemplateName, msg.TemplateParams,
		msg.WAMessageID, msg.Status, msg.ScheduledAt, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID returns a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT id, pending_line_id, client_id, direction, message_type,
			   content, media_url, template_name, template_params,
			   wa_message_id, wa_conversation_id, status, error_message,
			   scheduled_at, sent_at, delivered_at, read_at, created_at
		FROM messages
		WHERE id = $1
	`

	var msg models.Message
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&msg.ID, &msg.PendingLineID, &msg.ClientID, &msg.Direction, &msg.MessageType,
		&msg.Content, &msg.MediaURL, &msg.TemplateName, &msg.TemplateParams,
		&msg.WAMessageID, &msg.WAConversationID, &msg.Status, &msg.ErrorMessage,
		&msg.ScheduledAt, &msg.SentAt, &msg.DeliveredAt, &msg.ReadAt, &msg.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &msg, nil
}

// GetByWAMessageID returns a message by WhatsApp message ID
func (r *MessageRepository) GetByWAMessageID(ctx context.Context, waMessageID string) (*models.Message, error) {
	query := `
		SELECT id, pending_line_id, client_id, direction, message_type,
			   content, media_url, template_name, template_params,
			   wa_message_id, wa_conversation_id, status, error_message,
			   scheduled_at, sent_at, delivered_at, read_at, created_at
		FROM messages
		WHERE wa_message_id = $1
	`

	var msg models.Message
	err := r.pool.QueryRow(ctx, query, waMessageID).Scan(
		&msg.ID, &msg.PendingLineID, &msg.ClientID, &msg.Direction, &msg.MessageType,
		&msg.Content, &msg.MediaURL, &msg.TemplateName, &msg.TemplateParams,
		&msg.WAMessageID, &msg.WAConversationID, &msg.Status, &msg.ErrorMessage,
		&msg.ScheduledAt, &msg.SentAt, &msg.DeliveredAt, &msg.ReadAt, &msg.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message by WA ID: %w", err)
	}

	return &msg, nil
}

// ListByPendingLine returns all messages for a pending line
func (r *MessageRepository) ListByPendingLine(ctx context.Context, pendingLineID uuid.UUID) ([]models.Message, error) {
	query := `
		SELECT id, pending_line_id, client_id, direction, message_type,
			   content, media_url, template_name, template_params,
			   wa_message_id, wa_conversation_id, status, error_message,
			   scheduled_at, sent_at, delivered_at, read_at, created_at
		FROM messages
		WHERE pending_line_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, pendingLineID)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID, &msg.PendingLineID, &msg.ClientID, &msg.Direction, &msg.MessageType,
			&msg.Content, &msg.MediaURL, &msg.TemplateName, &msg.TemplateParams,
			&msg.WAMessageID, &msg.WAConversationID, &msg.Status, &msg.ErrorMessage,
			&msg.ScheduledAt, &msg.SentAt, &msg.DeliveredAt, &msg.ReadAt, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ListByClient returns all messages for a client
func (r *MessageRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, pending_line_id, client_id, direction, message_type,
			   content, media_url, template_name, template_params,
			   wa_message_id, wa_conversation_id, status, error_message,
			   scheduled_at, sent_at, delivered_at, read_at, created_at
		FROM messages
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID, &msg.PendingLineID, &msg.ClientID, &msg.Direction, &msg.MessageType,
			&msg.Content, &msg.MediaURL, &msg.TemplateName, &msg.TemplateParams,
			&msg.WAMessageID, &msg.WAConversationID, &msg.Status, &msg.ErrorMessage,
			&msg.ScheduledAt, &msg.SentAt, &msg.DeliveredAt, &msg.ReadAt, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateStatus updates the status and timestamps of a message
func (r *MessageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MessageStatus, waMessageID *string) error {
	query := `
		UPDATE messages SET 
			status = $2,
			wa_message_id = COALESCE($3, wa_message_id),
			sent_at = CASE WHEN $2 = 'sent' THEN NOW() ELSE sent_at END,
			delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END,
			read_at = CASE WHEN $2 = 'read' THEN NOW() ELSE read_at END
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, status, waMessageID)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// UpdateStatusByWAID updates message status using WhatsApp message ID
func (r *MessageRepository) UpdateStatusByWAID(ctx context.Context, waMessageID string, status models.MessageStatus) error {
	query := `
		UPDATE messages SET 
			status = $2,
			delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END,
			read_at = CASE WHEN $2 = 'read' THEN NOW() ELSE read_at END
		WHERE wa_message_id = $1
	`

	_, err := r.pool.Exec(ctx, query, waMessageID, status)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// SetError sets an error message on a failed message
func (r *MessageRepository) SetError(ctx context.Context, id uuid.UUID, errorMsg string) error {
	query := `UPDATE messages SET status = 'failed', error_message = $2 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to set error: %w", err)
	}

	return nil
}

// GetPendingCount returns the count of pending/queued messages
func (r *MessageRepository) GetPendingCount(ctx context.Context, cabinetID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM messages m
		JOIN pending_lines pl ON m.pending_line_id = pl.id
		WHERE pl.cabinet_id = $1 AND m.status IN ('queued', 'sending')
	`

	var count int
	err := r.pool.QueryRow(ctx, query, cabinetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending count: %w", err)
	}

	return count, nil
}
