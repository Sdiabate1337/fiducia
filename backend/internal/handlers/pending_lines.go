package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

// PendingLineHandler handles pending line requests
type PendingLineHandler struct {
	repo *repository.PendingLineRepository
}

// NewPendingLineHandler creates a new handler
func NewPendingLineHandler(repo *repository.PendingLineRepository) *PendingLineHandler {
	return &PendingLineHandler{repo: repo}
}

// List handles GET /api/v1/cabinets/{cabinet_id}/pending-lines
func (h *PendingLineHandler) List(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Build filter from query params
	filter := repository.PendingLineFilter{
		CabinetID: cabinetID,
	}

	// Parse query parameters
	query := r.URL.Query()

	if clientID := query.Get("client_id"); clientID != "" {
		if id, err := uuid.Parse(clientID); err == nil {
			filter.ClientID = &id
		}
	}

	if status := query.Get("status"); status != "" {
		s := models.PendingLineStatus(status)
		filter.Status = &s
	}

	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}

	if limit := query.Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			filter.Limit = n
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if n, err := strconv.Atoi(offset); err == nil {
			filter.Offset = n
		}
	}

	result, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list pending lines")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Get handles GET /api/v1/pending-lines/{id}
func (h *PendingLineHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	line, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get pending line")
		return
	}
	if line == nil {
		writeError(w, http.StatusNotFound, "Pending line not found")
		return
	}

	writeJSON(w, http.StatusOK, line)
}

// CreatePendingLineRequest represents the create request body
type CreatePendingLineRequest struct {
	ClientID        *uuid.UUID `json:"client_id,omitempty"`
	Amount          float64    `json:"amount"`
	TransactionDate string     `json:"transaction_date"` // YYYY-MM-DD
	BankLabel       *string    `json:"bank_label,omitempty"`
	AccountNumber   *string    `json:"account_number,omitempty"`
}

// Create handles POST /api/v1/cabinets/{cabinet_id}/pending-lines
func (h *PendingLineHandler) Create(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	var req CreatePendingLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Amount == 0 {
		writeError(w, http.StatusBadRequest, "Amount is required")
		return
	}
	if req.TransactionDate == "" {
		writeError(w, http.StatusBadRequest, "Transaction date is required")
		return
	}

	// Parse date
	date, err := parseDate(req.TransactionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
		return
	}

	line := &models.PendingLine{
		CabinetID:       cabinetID,
		ClientID:        req.ClientID,
		Amount:          decimalFromFloat(req.Amount),
		TransactionDate: date,
		BankLabel:       req.BankLabel,
		AccountNumber:   req.AccountNumber,
		Status:          models.StatusPending,
	}

	if err := h.repo.Create(r.Context(), line); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create pending line")
		return
	}

	writeJSON(w, http.StatusCreated, line)
}

// UpdatePendingLineRequest represents the update request body
type UpdatePendingLineRequest struct {
	ClientID   *uuid.UUID              `json:"client_id,omitempty"`
	Status     *models.PendingLineStatus `json:"status,omitempty"`
	AssignedTo *uuid.UUID              `json:"assigned_to,omitempty"`
}

// Update handles PATCH /api/v1/pending-lines/{id}
func (h *PendingLineHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	// Get existing line
	line, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get pending line")
		return
	}
	if line == nil {
		writeError(w, http.StatusNotFound, "Pending line not found")
		return
	}

	var req UpdatePendingLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	if req.ClientID != nil {
		line.ClientID = req.ClientID
	}
	if req.Status != nil {
		line.Status = *req.Status
	}
	if req.AssignedTo != nil {
		line.AssignedTo = req.AssignedTo
	}

	if err := h.repo.Update(r.Context(), line); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update pending line")
		return
	}

	writeJSON(w, http.StatusOK, line)
}

// Delete handles DELETE /api/v1/pending-lines/{id}
func (h *PendingLineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete pending line")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Stats handles GET /api/v1/cabinets/{cabinet_id}/pending-lines/stats
func (h *PendingLineHandler) Stats(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	stats, err := h.repo.GetStats(r.Context(), cabinetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
