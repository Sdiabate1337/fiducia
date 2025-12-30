package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

// ClientHandler handles client requests
type ClientHandler struct {
	repo *repository.ClientRepository
}

// NewClientHandler creates a new handler
func NewClientHandler(repo *repository.ClientRepository) *ClientHandler {
	return &ClientHandler{repo: repo}
}

// List handles GET /api/v1/cabinets/{cabinet_id}/clients
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	filter := repository.ClientFilter{
		CabinetID: cabinetID,
	}

	query := r.URL.Query()

	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}

	if phone := query.Get("phone"); phone != "" {
		filter.Phone = &phone
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
		writeError(w, http.StatusInternalServerError, "Failed to list clients")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Get handles GET /api/v1/clients/{id}
func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get client")
		return
	}
	if client == nil {
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}

	writeJSON(w, http.StatusOK, client)
}

// CreateClientRequest represents the create request body
type CreateClientRequest struct {
	Name        string  `json:"name"`
	SIREN       *string `json:"siren,omitempty"`
	SIRET       *string `json:"siret,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	ContactName *string `json:"contact_name,omitempty"`
	Address     *string `json:"address,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// Create handles POST /api/v1/cabinets/{cabinet_id}/clients
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	client := &models.Client{
		CabinetID:   cabinetID,
		Name:        req.Name,
		SIREN:       req.SIREN,
		SIRET:       req.SIRET,
		Phone:       req.Phone,
		Email:       req.Email,
		ContactName: req.ContactName,
		Address:     req.Address,
		Notes:       req.Notes,
	}

	if err := h.repo.Create(r.Context(), client); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create client")
		return
	}

	writeJSON(w, http.StatusCreated, client)
}

// UpdateClientRequest represents the update request body
type UpdateClientRequest struct {
	Name            *string `json:"name,omitempty"`
	SIREN           *string `json:"siren,omitempty"`
	SIRET           *string `json:"siret,omitempty"`
	Phone           *string `json:"phone,omitempty"`
	Email           *string `json:"email,omitempty"`
	ContactName     *string `json:"contact_name,omitempty"`
	Address         *string `json:"address,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	WhatsAppOptedIn *bool   `json:"whatsapp_opted_in,omitempty"`
}

// Update handles PUT /api/v1/clients/{id}
func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get client")
		return
	}
	if client == nil {
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.SIREN != nil {
		client.SIREN = req.SIREN
	}
	if req.SIRET != nil {
		client.SIRET = req.SIRET
	}
	if req.Phone != nil {
		client.Phone = req.Phone
	}
	if req.Email != nil {
		client.Email = req.Email
	}
	if req.ContactName != nil {
		client.ContactName = req.ContactName
	}
	if req.Address != nil {
		client.Address = req.Address
	}
	if req.Notes != nil {
		client.Notes = req.Notes
	}
	if req.WhatsAppOptedIn != nil {
		client.WhatsAppOptedIn = *req.WhatsAppOptedIn
	}

	if err := h.repo.Update(r.Context(), client); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update client")
		return
	}

	writeJSON(w, http.StatusOK, client)
}

// Delete handles DELETE /api/v1/clients/{id}
func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete client")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
