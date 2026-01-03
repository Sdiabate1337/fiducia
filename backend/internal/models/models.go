package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Cabinet represents an accounting firm (tenant)
type Cabinet struct {
	ID                  uuid.UUID      `json:"id"`
	Name                string         `json:"name"`
	SIRET               *string        `json:"siret,omitempty"`
	Email               *string        `json:"email,omitempty"`
	Phone               *string        `json:"phone,omitempty"`
	Address             *string        `json:"address,omitempty"`
	Settings            map[string]any `json:"settings"`
	OnboardingCompleted bool           `json:"onboarding_completed"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// Collaborator represents a cabinet employee
type Collaborator struct {
	ID             uuid.UUID `json:"id"`
	CabinetID      uuid.UUID `json:"cabinet_id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	VoiceID        *string   `json:"voice_id,omitempty"`
	VoiceSampleURL *string   `json:"voice_sample_url,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Client represents a cabinet's client
type Client struct {
	ID                uuid.UUID  `json:"id"`
	CabinetID         uuid.UUID  `json:"cabinet_id"`
	Name              string     `json:"name"`
	SIREN             *string    `json:"siren,omitempty"`
	SIRET             *string    `json:"siret,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	Email             *string    `json:"email,omitempty"`
	ContactName       *string    `json:"contact_name,omitempty"`
	Address           *string    `json:"address,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	WhatsAppOptedIn   bool       `json:"whatsapp_opted_in"`
	WhatsAppOptedInAt *time.Time `json:"whatsapp_opted_in_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// PendingLineStatus represents the status of a pending line
type PendingLineStatus string

const (
	StatusPending   PendingLineStatus = "pending"
	StatusContacted PendingLineStatus = "contacted"
	StatusReceived  PendingLineStatus = "received"
	StatusValidated PendingLineStatus = "validated"
	StatusRejected  PendingLineStatus = "rejected"
	StatusExpired   PendingLineStatus = "expired"
)

// PendingLine represents a line in compte 471
type PendingLine struct {
	ID              uuid.UUID         `json:"id"`
	CabinetID       uuid.UUID         `json:"cabinet_id"`
	ClientID        *uuid.UUID        `json:"client_id,omitempty"`
	Amount          decimal.Decimal   `json:"amount"`
	TransactionDate time.Time         `json:"transaction_date"`
	BankLabel       *string           `json:"bank_label,omitempty"`
	AccountNumber   *string           `json:"account_number,omitempty"`
	ImportBatchID   *uuid.UUID        `json:"import_batch_id,omitempty"`
	SourceFile      *string           `json:"source_file,omitempty"`
	SourceRowNumber *int              `json:"source_row_number,omitempty"`
	Status          PendingLineStatus `json:"status"`
	LastContactedAt *time.Time        `json:"last_contacted_at,omitempty"`
	ContactCount    int               `json:"contact_count"`
	AssignedTo      *uuid.UUID        `json:"assigned_to,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`

	// Relations (populated via joins)
	Client *Client `json:"client,omitempty"`

	// Enriched Fields (Campaign Status)
	CampaignStatus      *string    `json:"campaign_status,omitempty"`
	NextStepScheduledAt *time.Time `json:"next_step_scheduled_at,omitempty"`
	CampaignCurrentStep *int       `json:"campaign_current_step,omitempty"`
}

// ImportBatch represents a CSV import batch
type ImportBatch struct {
	ID           uuid.UUID      `json:"id"`
	CabinetID    uuid.UUID      `json:"cabinet_id"`
	ImportedBy   *uuid.UUID     `json:"imported_by,omitempty"`
	Filename     *string        `json:"filename,omitempty"`
	FileType     *string        `json:"file_type,omitempty"`
	TotalRows    *int           `json:"total_rows,omitempty"`
	ImportedRows *int           `json:"imported_rows,omitempty"`
	FailedRows   *int           `json:"failed_rows,omitempty"`
	Errors       map[string]any `json:"errors,omitempty"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}

// MessageDirection represents the direction of a message
type MessageDirection string

const (
	DirectionOutbound MessageDirection = "outbound"
	DirectionInbound  MessageDirection = "inbound"
)

// MessageType represents the type of message
type MessageType string

const (
	TypeText        MessageType = "text"
	TypeVoice       MessageType = "voice"
	TypeInteractive MessageType = "interactive"
	TypeMedia       MessageType = "media"
	TypeTemplate    MessageType = "template"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MsgStatusQueued    MessageStatus = "queued"
	MsgStatusSending   MessageStatus = "sending"
	MsgStatusSent      MessageStatus = "sent"
	MsgStatusDelivered MessageStatus = "delivered"
	MsgStatusRead      MessageStatus = "read"
	MsgStatusFailed    MessageStatus = "failed"
	MsgStatusReceived  MessageStatus = "received"
)

// Message represents a WhatsApp message
type Message struct {
	ID               uuid.UUID        `json:"id"`
	PendingLineID    *uuid.UUID       `json:"pending_line_id,omitempty"`
	ClientID         *uuid.UUID       `json:"client_id,omitempty"`
	Direction        MessageDirection `json:"direction"`
	MessageType      MessageType      `json:"message_type"`
	Content          *string          `json:"content,omitempty"`
	MediaURL         *string          `json:"media_url,omitempty"`
	TemplateName     *string          `json:"template_name,omitempty"`
	TemplateParams   map[string]any   `json:"template_params,omitempty"`
	WAMessageID      *string          `json:"wa_message_id,omitempty"`
	WAConversationID *string          `json:"wa_conversation_id,omitempty"`
	Status           MessageStatus    `json:"status"`
	ErrorMessage     *string          `json:"error_message,omitempty"`
	ScheduledAt      *time.Time       `json:"scheduled_at,omitempty"`
	SentAt           *time.Time       `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time       `json:"delivered_at,omitempty"`
	ReadAt           *time.Time       `json:"read_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// DocumentType represents the type of document
type DocumentType string

const (
	DocTypeReceipt       DocumentType = "receipt"
	DocTypeInvoice       DocumentType = "invoice"
	DocTypeBankStatement DocumentType = "bank_statement"
	DocTypeContract      DocumentType = "contract"
	DocTypeOther         DocumentType = "other"
)

// ReceivedDocument represents a document received from a client
type ReceivedDocument struct {
	ID                uuid.UUID        `json:"id"`
	PendingLineID     *uuid.UUID       `json:"pending_line_id,omitempty"`
	MessageID         *uuid.UUID       `json:"message_id,omitempty"`
	ClientID          *uuid.UUID       `json:"client_id,omitempty"`
	FileURL           string           `json:"file_url"`
	FileType          *string          `json:"file_type,omitempty"`
	FileSize          *int             `json:"file_size,omitempty"`
	OriginalFilename  *string          `json:"original_filename,omitempty"`
	DocumentType      *DocumentType    `json:"document_type,omitempty"`
	OCRResult         map[string]any   `json:"ocr_result,omitempty"`
	OCRConfidence     *decimal.Decimal `json:"ocr_confidence,omitempty"`
	ExtractedAmount   *decimal.Decimal `json:"extracted_amount,omitempty"`
	ExtractedDate     *time.Time       `json:"extracted_date,omitempty"`
	ExtractedMerchant *string          `json:"extracted_merchant,omitempty"`
	OCRStatus         string           `json:"ocr_status"`
	OCRError          *string          `json:"ocr_error,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
}

// ProposalStatus represents the status of a matching proposal
type ProposalStatus string

const (
	ProposalPending  ProposalStatus = "pending"
	ProposalApproved ProposalStatus = "approved"
	ProposalRejected ProposalStatus = "rejected"
)

// MatchingProposal represents a proposed match between a line and document
type MatchingProposal struct {
	ID              uuid.UUID        `json:"id"`
	PendingLineID   uuid.UUID        `json:"pending_line_id"`
	DocumentID      uuid.UUID        `json:"document_id"`
	ProposedBy      string           `json:"proposed_by"`
	MatchConfidence *decimal.Decimal `json:"match_confidence,omitempty"`
	MatchReasons    map[string]any   `json:"match_reasons,omitempty"`
	Status          ProposalStatus   `json:"status"`
	ValidatedBy     *uuid.UUID       `json:"validated_by,omitempty"`
	ValidatedAt     *time.Time       `json:"validated_at,omitempty"`
	RejectionReason *string          `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`

	// Relations (populated via joins)
	PendingLine *PendingLine      `json:"pending_line,omitempty"`
	Document    *ReceivedDocument `json:"document,omitempty"`
}
