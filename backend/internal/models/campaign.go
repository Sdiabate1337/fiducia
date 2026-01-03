package models

import (
	"time"

	"github.com/google/uuid"
)

// CampaignTriggerType represents when a campaign starts
type CampaignTriggerType string

const (
	TriggerOnPending CampaignTriggerType = "on_pending"
	TriggerOnOverdue CampaignTriggerType = "on_overdue"
)

// CampaignChannel represents the delivery channel
type CampaignChannel string

const (
	ChannelWhatsApp     CampaignChannel = "whatsapp"
	ChannelEmail        CampaignChannel = "email"
	ChannelVoice        CampaignChannel = "voice"
	ChannelNotification CampaignChannel = "notification"
)

// ExecutionStatus represents the state of a campaign execution
type ExecutionStatus string

const (
	ExecStatusPending   ExecutionStatus = "pending"
	ExecStatusRunning   ExecutionStatus = "running"
	ExecStatusCompleted ExecutionStatus = "completed"
	ExecStatusStopped   ExecutionStatus = "stopped"
	ExecStatusFailed    ExecutionStatus = "failed"
)

// StopReason represents why a campaign was stopped
type StopReason string

const (
	StopOCRValidated    StopReason = "ocr_validated"
	StopManualValidated StopReason = "manual_validaton"
	StopClientRefusal   StopReason = "client_refusal"
	StopCompleted       StopReason = "completed"
)

// Campaign represents a sequence of automated actions
type Campaign struct {
	ID                uuid.UUID           `json:"id"`
	CabinetID         uuid.UUID           `json:"cabinet_id"`
	Name              string              `json:"name"`
	TriggerType       CampaignTriggerType `json:"trigger_type"`
	IsActive          bool                `json:"is_active"`
	QuietHoursEnabled bool                `json:"quiet_hours_enabled"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`

	// Relations
	Steps []CampaignStep `json:"steps,omitempty"`
}

// CampaignStep represents a single action in the sequence
type CampaignStep struct {
	ID         uuid.UUID       `json:"id"`
	CampaignID uuid.UUID       `json:"campaign_id"`
	StepOrder  int             `json:"step_order"`
	DelayHours int             `json:"delay_hours"`
	Channel    CampaignChannel `json:"channel"`
	TemplateID string          `json:"template_id"`
	Config     map[string]any  `json:"config,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// CampaignExecution tracks the progress of a campaign for a specific pending line
type CampaignExecution struct {
	ID                  uuid.UUID       `json:"id"`
	CampaignID          uuid.UUID       `json:"campaign_id"`
	PendingLineID       uuid.UUID       `json:"pending_line_id"`
	CurrentStepOrder    int             `json:"current_step_order"`
	Status              ExecutionStatus `json:"status"`
	StopReason          *StopReason     `json:"stop_reason,omitempty"`
	LastStepExecutedAt  *time.Time      `json:"last_step_executed_at,omitempty"`
	NextStepScheduledAt *time.Time      `json:"next_step_scheduled_at,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}
