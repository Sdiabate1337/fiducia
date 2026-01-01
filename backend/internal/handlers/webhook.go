package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
	"github.com/fiducia/backend/internal/services"
)

// WebhookHandler handles incoming WhatsApp webhooks
type WebhookHandler struct {
	cfg         *config.Config
	pool        *pgxpool.Pool
	ocrSvc      *services.OCRService
	matchingSvc *services.MatchingService
	docRepo     *repository.DocumentRepository
	clientRepo  *repository.ClientRepository
	msgRepo     *repository.MessageRepository
	lineRepo    *repository.PendingLineRepository
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	cfg *config.Config,
	pool *pgxpool.Pool,
	ocrSvc *services.OCRService,
	matchingSvc *services.MatchingService,
	docRepo *repository.DocumentRepository,
	clientRepo *repository.ClientRepository,
	msgRepo *repository.MessageRepository,
	lineRepo *repository.PendingLineRepository,
) *WebhookHandler {
	return &WebhookHandler{
		cfg:         cfg,
		pool:        pool,
		ocrSvc:      ocrSvc,
		matchingSvc: matchingSvc,
		docRepo:     docRepo,
		clientRepo:  clientRepo,
		msgRepo:     msgRepo,
		lineRepo:    lineRepo,
	}
}

// TwilioWebhookPayload represents the Twilio webhook payload
type TwilioWebhookPayload struct {
	MessageSid        string `json:"MessageSid"`
	AccountSid        string `json:"AccountSid"`
	From              string `json:"From"`
	To                string `json:"To"`
	Body              string `json:"Body"`
	NumMedia          string `json:"NumMedia"`
	MediaContentType0 string `json:"MediaContentType0,omitempty"`
	MediaUrl0         string `json:"MediaUrl0,omitempty"`
	Status            string `json:"MessageStatus,omitempty"` // For status callbacks
	ErrorCode         string `json:"ErrorCode,omitempty"`
	ErrorMessage      string `json:"ErrorMessage,omitempty"`
}

// StatusCallbackPayload for message status updates
type StatusCallbackPayload struct {
	MessageSid    string `json:"MessageSid"`
	MessageStatus string `json:"MessageStatus"`
	To            string `json:"To"`
	ErrorCode     string `json:"ErrorCode,omitempty"`
	ErrorMessage  string `json:"ErrorMessage,omitempty"`
}

// HandleIncoming handles incoming WhatsApp messages
func (h *WebhookHandler) HandleIncoming(w http.ResponseWriter, r *http.Request) {
	// Verify Twilio signature
	if !h.verifyTwilioSignature(r) {
		slog.Warn("invalid Twilio signature")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse webhook form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	payload := TwilioWebhookPayload{
		MessageSid:        r.FormValue("MessageSid"),
		AccountSid:        r.FormValue("AccountSid"),
		From:              r.FormValue("From"),
		To:                r.FormValue("To"),
		Body:              r.FormValue("Body"),
		NumMedia:          r.FormValue("NumMedia"),
		MediaContentType0: r.FormValue("MediaContentType0"),
		MediaUrl0:         r.FormValue("MediaUrl0"),
	}

	// Log incoming message
	slog.Info("received WhatsApp message",
		"message_sid", payload.MessageSid,
		"from", payload.From,
		"body_length", len(payload.Body),
		"num_media", payload.NumMedia,
	)

	// Clean phone number (remove whatsapp: prefix)
	from := strings.TrimPrefix(payload.From, "whatsapp:")

	// Find client by phone number
	client, err := h.clientRepo.GetByPhoneGlobal(r.Context(), from)
	if err != nil {
		slog.Warn("client not found for phone", "phone", from, "error", err)
	}

	var clientID *uuid.UUID
	if client != nil {
		clientID = &client.ID
	}

	// Save incoming message to database
	waMessageID := payload.MessageSid
	msg := &models.Message{
		ClientID:    clientID,
		Direction:   models.DirectionInbound,
		MessageType: models.TypeText,
		Content:     &payload.Body,
		Status:      models.MsgStatusDelivered,
		WAMessageID: &waMessageID,
	}

	if payload.NumMedia != "0" {
		msg.MessageType = models.TypeMedia
	}

	if err := h.msgRepo.Create(r.Context(), msg); err != nil {
		slog.Error("failed to save incoming message", "error", err)
	}

	// Handle media if present
	if payload.NumMedia != "0" && payload.MediaUrl0 != "" {
		slog.Info("message has media attachment",
			"content_type", payload.MediaContentType0,
			"url", payload.MediaUrl0,
		)

		// Process media asynchronously
		go h.processMedia(r.Context(), payload, clientID, msg.ID)
	}

	// Acknowledge receipt (Twilio expects 200 OK)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	// Empty TwiML response
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Response></Response>`)

	// Log for tracking
	slog.Info("webhook processed successfully",
		"from", from,
		"message_id", payload.MessageSid,
	)
}

// processMedia downloads and processes media with OCR
func (h *WebhookHandler) processMedia(ctx context.Context, payload TwilioWebhookPayload, clientID *uuid.UUID, messageID uuid.UUID) {
	// Download from Twilio and process with OCR
	ocrResult, filePath, err := h.ocrSvc.DownloadAndProcess(
		ctx,
		payload.MediaUrl0,
		h.cfg.TwilioAccountSID,
		h.cfg.TwilioAuthToken,
	)

	// Create document record
	doc := &repository.Document{
		ClientID:        clientID,
		MessageID:       &messageID,
		FilePath:        filePath,
		FileName:        strPtr(filepath.Base(filePath)),
		FileType:        strPtr(payload.MediaContentType0),
		TwilioMediaURL:  strPtr(payload.MediaUrl0),
		MatchConfidence: decimal.Zero,
		MatchStatus:     "pending",
	}

	if err != nil {
		slog.Error("OCR processing failed", "error", err)
		doc.OCRStatus = "failed"
		doc.OCRError = strPtr(err.Error())
	} else {
		doc.OCRStatus = "completed"
		doc.OCRText = strPtr(ocrResult.RawText)
		doc.OCRData = ocrResult.ExtractedData
	}

	if err := h.docRepo.Create(ctx, doc); err != nil {
		slog.Error("failed to save document", "error", err)
		return
	}

	slog.Info("document saved",
		"doc_id", doc.ID,
		"ocr_status", doc.OCRStatus,
		"client_id", clientID,
	)

	// If OCR succeeded, attempt auto-matching
	if doc.OCRStatus == "completed" && clientID != nil {
		proposal, err := h.matchingSvc.AutoMatch(ctx, doc)
		if err != nil {
			slog.Error("auto-match failed", "error", err)
		} else if proposal != nil {
			slog.Info("match proposal created",
				"doc_id", doc.ID,
				"line_id", proposal.PendingLineID,
				"confidence", proposal.Confidence,
			)
		}
	}
}

// HandleStatusCallback handles message status updates
func (h *WebhookHandler) HandleStatusCallback(w http.ResponseWriter, r *http.Request) {
	// Verify Twilio signature
	if !h.verifyTwilioSignature(r) {
		slog.Warn("invalid Twilio signature on status callback")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse status callback form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	payload := StatusCallbackPayload{
		MessageSid:    r.FormValue("MessageSid"),
		MessageStatus: r.FormValue("MessageStatus"),
		To:            r.FormValue("To"),
		ErrorCode:     r.FormValue("ErrorCode"),
		ErrorMessage:  r.FormValue("ErrorMessage"),
	}

	slog.Info("received status callback",
		"message_sid", payload.MessageSid,
		"status", payload.MessageStatus,
		"error_code", payload.ErrorCode,
	)

	// Map Twilio status to our status
	var status models.MessageStatus
	switch payload.MessageStatus {
	case "queued":
		status = models.MsgStatusQueued
	case "sending":
		status = models.MsgStatusSending
	case "sent":
		status = models.MsgStatusSent
	case "delivered":
		status = models.MsgStatusDelivered
	case "read":
		status = models.MsgStatusRead
	case "failed", "undelivered":
		status = models.MsgStatusFailed
	default:
		status = models.MsgStatusQueued
	}

	// Update message status in database
	if err := h.msgRepo.UpdateStatusByWAID(r.Context(), payload.MessageSid, status); err != nil {
		slog.Error("failed to update message status", "error", err)
	}

	// Acknowledge
	w.WriteHeader(http.StatusOK)
}

// verifyTwilioSignature validates the X-Twilio-Signature header
func (h *WebhookHandler) verifyTwilioSignature(r *http.Request) bool {
	// In development, skip verification
	if h.cfg.Environment == "development" {
		return true
	}

	signature := r.Header.Get("X-Twilio-Signature")
	if signature == "" {
		return false
	}

	// Build the full URL
	url := "https://" + r.Host + r.URL.Path

	// Get POST parameters sorted and concatenated
	var params []string
	for key, values := range r.Form {
		for _, value := range values {
			params = append(params, key+value)
		}
	}

	// Create HMAC-SHA1 signature
	mac := hmac.New(sha256.New, []byte(h.cfg.TwilioAuthToken))
	mac.Write([]byte(url))
	for _, param := range params {
		mac.Write([]byte(param))
	}
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// WebhookResponse for JSON API responses
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// HandleIncomingJSON handles incoming messages as JSON (for testing)
func (h *WebhookHandler) HandleIncomingJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read body")
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	slog.Info("received JSON webhook",
		"payload", payload,
	)

	writeJSON(w, http.StatusOK, WebhookResponse{
		Success: true,
		Message: "Webhook received",
	})
}

func strPtr(s string) *string {
	return &s
}
