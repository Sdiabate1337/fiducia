package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/queue"
	"github.com/fiducia/backend/internal/repository"
	"github.com/fiducia/backend/pkg/whatsapp"
)

// MessageService handles sending WhatsApp messages
type MessageService struct {
	waClient   whatsapp.Client
	msgRepo    *repository.MessageRepository
	lineRepo   *repository.PendingLineRepository
	clientRepo *repository.ClientRepository
	queue      *queue.MessageQueue
}

// NewMessageService creates a new message service
func NewMessageService(
	waClient whatsapp.Client,
	msgRepo *repository.MessageRepository,
	lineRepo *repository.PendingLineRepository,
	clientRepo *repository.ClientRepository,
	q *queue.MessageQueue,
) *MessageService {
	return &MessageService{
		waClient:   waClient,
		msgRepo:    msgRepo,
		lineRepo:   lineRepo,
		clientRepo: clientRepo,
		queue:      q,
	}
}

// SendRelanceRequest represents a request to send a relance
type SendRelanceRequest struct {
	PendingLineID uuid.UUID `json:"pending_line_id"`
	MessageType   string    `json:"message_type"` // text, voice, template
	CustomMessage string    `json:"custom_message,omitempty"`
	Immediate     bool      `json:"immediate,omitempty"` // Skip queue for testing
}

// SendRelance queues a relance message for a pending line
func (s *MessageService) SendRelance(ctx context.Context, req SendRelanceRequest) (*models.Message, error) {
	// Get pending line
	line, err := s.lineRepo.GetByID(ctx, req.PendingLineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending line: %w", err)
	}
	if line == nil {
		return nil, fmt.Errorf("pending line not found")
	}

	// Check if client is assigned
	if line.ClientID == nil {
		return nil, fmt.Errorf("no client assigned to this pending line")
	}

	// Get client
	client, err := s.clientRepo.GetByID(ctx, *line.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("client not found")
	}

	// Check phone number
	if client.Phone == nil || *client.Phone == "" {
		return nil, fmt.Errorf("client has no phone number")
	}

	// Generate message content
	content := req.CustomMessage
	if content == "" {
		content = s.generateRelanceMessage(line, client)
	}

	// Create message record
	msg := &models.Message{
		ID:            uuid.New(),
		PendingLineID: &req.PendingLineID,
		ClientID:      line.ClientID,
		Direction:     models.DirectionOutbound,
		MessageType:   models.MessageType(req.MessageType),
		Content:       &content,
		Status:        models.MsgStatusQueued,
	}

	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Create queue job
	job := &queue.MessageJob{
		ID:            msg.ID.String(),
		PendingLineID: req.PendingLineID.String(),
		ClientID:      line.ClientID.String(),
		Phone:         *client.Phone,
		MessageType:   req.MessageType,
		Content:       content,
	}

	// Enqueue (with or without jitter)
	if req.Immediate {
		if err := s.queue.EnqueueImmediate(ctx, job); err != nil {
			return nil, fmt.Errorf("failed to enqueue message: %w", err)
		}
	} else {
		if err := s.queue.Enqueue(ctx, job); err != nil {
			return nil, fmt.Errorf("failed to enqueue message: %w", err)
		}
	}

	// Update pending line status
	line.Status = models.StatusContacted
	line.ContactCount++
	now := time.Now()
	line.LastContactedAt = &now
	if err := s.lineRepo.Update(ctx, line); err != nil {
		slog.Warn("failed to update pending line status", "error", err)
	}

	return msg, nil
}

// ProcessQueuedMessage processes a message from the queue
func (s *MessageService) ProcessQueuedMessage(ctx context.Context, job *queue.MessageJob) error {
	msgID, err := uuid.Parse(job.ID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	// Update status to sending
	if err := s.msgRepo.UpdateStatus(ctx, msgID, models.MsgStatusSending, nil); err != nil {
		slog.Warn("failed to update message status to sending", "error", err)
	}

	// Simulate typing delay (anti-ban)
	time.Sleep(2 * time.Second)

	// Send via Twilio
	var response *whatsapp.MessageResponse
	var sendErr error

	switch job.MessageType {
	case "text":
		response, sendErr = s.waClient.SendText(job.Phone, job.Content)
	case "voice":
		response, sendErr = s.waClient.SendVoice(job.Phone, job.AudioURL)
	case "template":
		response, sendErr = s.waClient.SendTemplate(job.Phone, job.TemplateName, job.TemplateParams)
	default:
		response, sendErr = s.waClient.SendText(job.Phone, job.Content)
	}

	if sendErr != nil {
		// Mark as failed
		s.msgRepo.SetError(ctx, msgID, sendErr.Error())
		return fmt.Errorf("failed to send message: %w", sendErr)
	}

	// Update with WhatsApp message ID
	waID := response.MessageSID
	if err := s.msgRepo.UpdateStatus(ctx, msgID, models.MsgStatusSent, &waID); err != nil {
		slog.Warn("failed to update message status to sent", "error", err)
	}

	slog.Info("message sent successfully",
		"message_id", msgID,
		"wa_message_id", waID,
		"phone", job.Phone,
	)

	return nil
}

// generateRelanceMessage generates a default relance message
func (s *MessageService) generateRelanceMessage(line *models.PendingLine, client *models.Client) string {
	// Format amount
	amount := line.Amount.StringFixed(2)

	// Format date
	date := line.TransactionDate.Format("02/01/2006")

	// Get label
	label := "une opération"
	if line.BankLabel != nil {
		label = *line.BankLabel
	}

	return fmt.Sprintf(
		"Bonjour %s,\n\n"+
			"Nous recherchons un justificatif pour l'opération suivante :\n\n"+
			"📅 Date : %s\n"+
			"💰 Montant : %s €\n"+
			"📝 Libellé : %s\n\n"+
			"Merci de nous envoyer la pièce justificative (facture, ticket, reçu).\n\n"+
			"Cordialement,\n"+
			"Votre cabinet comptable",
		client.Name, date, amount, label,
	)
}

// HandleIncomingMessage processes an incoming WhatsApp message
func (s *MessageService) HandleIncomingMessage(ctx context.Context, from string, body string, mediaURL *string) error {
	// TODO: Implement incoming message handling
	// - Find client by phone
	// - Find recent pending lines for this client
	// - Create message record
	// - If media, trigger OCR processing

	slog.Info("received incoming message",
		"from", from,
		"body_length", len(body),
		"has_media", mediaURL != nil,
	)

	return nil
}
