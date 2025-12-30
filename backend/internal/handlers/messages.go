package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/repository"
	"github.com/fiducia/backend/internal/services"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	msgRepo  *repository.MessageRepository
	lineRepo *repository.PendingLineRepository
	msgSvc   *services.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(pool *pgxpool.Pool, msgSvc *services.MessageService) *MessageHandler {
	return &MessageHandler{
		msgRepo:  repository.NewMessageRepository(pool),
		lineRepo: repository.NewPendingLineRepository(pool),
		msgSvc:   msgSvc,
	}
}

// List handles GET /api/v1/pending-lines/{id}/messages
func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	messages, err := h.msgRepo.ListByPendingLine(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list messages")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"total":    len(messages),
	})
}

// SendRelanceRequest represents the request to send a relance
type SendRelanceRequest struct {
	MessageType   string `json:"message_type"` // text, voice, template
	CustomMessage string `json:"custom_message,omitempty"`
	Immediate     bool   `json:"immediate,omitempty"`
}

// Send handles POST /api/v1/pending-lines/{id}/messages
func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	pendingLineID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	var req SendRelanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to text message
		req = SendRelanceRequest{MessageType: "text"}
	}

	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Check if message service is available
	if h.msgSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "Message service not configured")
		return
	}

	msg, err := h.msgSvc.SendRelance(r.Context(), services.SendRelanceRequest{
		PendingLineID: pendingLineID,
		MessageType:   req.MessageType,
		CustomMessage: req.CustomMessage,
		Immediate:     req.Immediate,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Relance queued",
		"id":      msg.ID,
		"status":  msg.Status,
	})
}

// GetByID handles GET /api/v1/messages/{id}
func (h *MessageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	msg, err := h.msgRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get message")
		return
	}
	if msg == nil {
		writeError(w, http.StatusNotFound, "Message not found")
		return
	}

	writeJSON(w, http.StatusOK, msg)
}
