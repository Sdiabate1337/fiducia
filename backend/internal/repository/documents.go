package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Document represents a received document with OCR data
type Document struct {
	ID              uuid.UUID              `json:"id"`
	PendingLineID   *uuid.UUID             `json:"pending_line_id,omitempty"`
	ClientID        *uuid.UUID             `json:"client_id,omitempty"`
	MessageID       *uuid.UUID             `json:"message_id,omitempty"`
	FilePath        string                 `json:"file_path"`
	FileName        *string                `json:"file_name,omitempty"`
	FileType        *string                `json:"file_type,omitempty"`
	FileSize        *int                   `json:"file_size,omitempty"`
	TwilioMediaURL  *string                `json:"twilio_media_url,omitempty"`
	OCRText         *string                `json:"ocr_text,omitempty"`
	OCRData         map[string]interface{} `json:"ocr_data,omitempty"`
	OCRStatus       string                 `json:"ocr_status"`
	OCRError        *string                `json:"ocr_error,omitempty"`
	MatchConfidence decimal.Decimal        `json:"match_confidence"`
	MatchStatus     string                 `json:"match_status"`
	MatchedBy       *uuid.UUID             `json:"matched_by,omitempty"`
	MatchedAt       *time.Time             `json:"matched_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// DocumentRepository handles document operations
type DocumentRepository struct {
	pool *pgxpool.Pool
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

// Create inserts a new document
func (r *DocumentRepository) Create(ctx context.Context, doc *Document) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}

	ocrDataJSON, _ := json.Marshal(doc.OCRData)

	query := `
		INSERT INTO documents (
			id, pending_line_id, client_id, message_id,
			file_path, file_name, file_type, file_size, twilio_media_url,
			ocr_text, ocr_data, ocr_status, ocr_error,
			match_confidence, match_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		doc.ID, doc.PendingLineID, doc.ClientID, doc.MessageID,
		doc.FilePath, doc.FileName, doc.FileType, doc.FileSize, doc.TwilioMediaURL,
		doc.OCRText, ocrDataJSON, doc.OCRStatus, doc.OCRError,
		doc.MatchConfidence, doc.MatchStatus,
	).Scan(&doc.CreatedAt, &doc.UpdatedAt)
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	query := `
		SELECT id, pending_line_id, client_id, message_id,
			file_path, file_name, file_type, file_size, twilio_media_url,
			ocr_text, ocr_data, ocr_status, ocr_error,
			match_confidence, match_status, matched_by, matched_at,
			created_at, updated_at
		FROM documents WHERE id = $1
	`

	var doc Document
	var ocrDataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.PendingLineID, &doc.ClientID, &doc.MessageID,
		&doc.FilePath, &doc.FileName, &doc.FileType, &doc.FileSize, &doc.TwilioMediaURL,
		&doc.OCRText, &ocrDataJSON, &doc.OCRStatus, &doc.OCRError,
		&doc.MatchConfidence, &doc.MatchStatus, &doc.MatchedBy, &doc.MatchedAt,
		&doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(ocrDataJSON) > 0 {
		json.Unmarshal(ocrDataJSON, &doc.OCRData)
	}

	return &doc, nil
}

// GetByPendingLine retrieves documents for a pending line
func (r *DocumentRepository) GetByPendingLine(ctx context.Context, pendingLineID uuid.UUID) ([]*Document, error) {
	query := `
		SELECT id, pending_line_id, client_id, message_id,
			file_path, file_name, file_type, file_size, twilio_media_url,
			ocr_text, ocr_data, ocr_status, ocr_error,
			match_confidence, match_status, matched_by, matched_at,
			created_at, updated_at
		FROM documents 
		WHERE pending_line_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, pendingLineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		var doc Document
		var ocrDataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.PendingLineID, &doc.ClientID, &doc.MessageID,
			&doc.FilePath, &doc.FileName, &doc.FileType, &doc.FileSize, &doc.TwilioMediaURL,
			&doc.OCRText, &ocrDataJSON, &doc.OCRStatus, &doc.OCRError,
			&doc.MatchConfidence, &doc.MatchStatus, &doc.MatchedBy, &doc.MatchedAt,
			&doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(ocrDataJSON) > 0 {
			json.Unmarshal(ocrDataJSON, &doc.OCRData)
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// GetByClient retrieves documents for a client
func (r *DocumentRepository) GetByClient(ctx context.Context, clientID uuid.UUID) ([]*Document, error) {
	query := `
		SELECT id, pending_line_id, client_id, message_id,
			file_path, file_name, file_type, file_size, twilio_media_url,
			ocr_text, ocr_data, ocr_status, ocr_error,
			match_confidence, match_status, matched_by, matched_at,
			created_at, updated_at
		FROM documents 
		WHERE client_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		var doc Document
		var ocrDataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.PendingLineID, &doc.ClientID, &doc.MessageID,
			&doc.FilePath, &doc.FileName, &doc.FileType, &doc.FileSize, &doc.TwilioMediaURL,
			&doc.OCRText, &ocrDataJSON, &doc.OCRStatus, &doc.OCRError,
			&doc.MatchConfidence, &doc.MatchStatus, &doc.MatchedBy, &doc.MatchedAt,
			&doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(ocrDataJSON) > 0 {
			json.Unmarshal(ocrDataJSON, &doc.OCRData)
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// GetUnmatched retrieves documents pending matching
func (r *DocumentRepository) GetUnmatched(ctx context.Context) ([]*Document, error) {
	query := `
		SELECT id, pending_line_id, client_id, message_id,
			file_path, file_name, file_type, file_size, twilio_media_url,
			ocr_text, ocr_data, ocr_status, ocr_error,
			match_confidence, match_status, matched_by, matched_at,
			created_at, updated_at
		FROM documents 
		WHERE match_status = 'pending' AND ocr_status = 'completed'
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		var doc Document
		var ocrDataJSON []byte

		if err := rows.Scan(
			&doc.ID, &doc.PendingLineID, &doc.ClientID, &doc.MessageID,
			&doc.FilePath, &doc.FileName, &doc.FileType, &doc.FileSize, &doc.TwilioMediaURL,
			&doc.OCRText, &ocrDataJSON, &doc.OCRStatus, &doc.OCRError,
			&doc.MatchConfidence, &doc.MatchStatus, &doc.MatchedBy, &doc.MatchedAt,
			&doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(ocrDataJSON) > 0 {
			json.Unmarshal(ocrDataJSON, &doc.OCRData)
		}

		docs = append(docs, &doc)
	}

	return docs, nil
}

// UpdateOCRResult updates OCR processing results
func (r *DocumentRepository) UpdateOCRResult(ctx context.Context, id uuid.UUID, ocrText string, ocrData map[string]interface{}, status string, ocrError *string) error {
	ocrDataJSON, _ := json.Marshal(ocrData)

	query := `
		UPDATE documents 
		SET ocr_text = $2, ocr_data = $3, ocr_status = $4, ocr_error = $5
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, ocrText, ocrDataJSON, status, ocrError)
	return err
}

// UpdateMatch updates matching results
func (r *DocumentRepository) UpdateMatch(ctx context.Context, id uuid.UUID, pendingLineID *uuid.UUID, confidence decimal.Decimal, status string) error {
	query := `
		UPDATE documents 
		SET pending_line_id = $2, match_confidence = $3, match_status = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, pendingLineID, confidence, status)
	return err
}

// ApproveMatch approves a document match
func (r *DocumentRepository) ApproveMatch(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE documents 
		SET match_status = 'approved', matched_by = $2, matched_at = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, userID, now)
	return err
}

// RejectMatch rejects a document match
func (r *DocumentRepository) RejectMatch(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE documents 
		SET match_status = 'rejected', matched_by = $2, matched_at = $3, pending_line_id = NULL
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, userID, now)
	return err
}
