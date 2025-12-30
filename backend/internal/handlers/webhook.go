package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fiducia/backend/internal/config"
)

// WebhookHandler handles incoming WhatsApp webhooks
type WebhookHandler struct {
	cfg *config.Config
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{cfg: cfg}
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

	// TODO: Process incoming message
	// 1. Find client by phone number
	// 2. Find related pending lines
	// 3. Save message to database
	// 4. If media attached, trigger OCR processing

	// Handle media if present
	if payload.NumMedia != "0" && payload.MediaUrl0 != "" {
		slog.Info("message has media attachment",
			"content_type", payload.MediaContentType0,
			"url", payload.MediaUrl0,
		)
		// TODO: Download media and process with OCR
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

	// TODO: Update message status in database
	// - queued, sending, sent, delivered, read, failed

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
