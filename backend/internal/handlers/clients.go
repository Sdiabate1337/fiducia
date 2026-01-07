package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

// listClients handles GET /api/v1/cabinets/{cabinet_id}/clients
func (r *Router) listClients(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	filter := repository.ClientFilter{
		CabinetID: cabinetID,
	}

	query := req.URL.Query()

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

	result, err := r.clientRepo.List(req.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list clients")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// getClient handles GET /api/v1/clients/{id}
func (r *Router) getClient(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := r.clientRepo.GetByID(req.Context(), id)
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

// createClient handles POST /api/v1/cabinets/{cabinet_id}/clients
func (r *Router) createClient(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	var payload CreateClientRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	client := &models.Client{
		CabinetID:   cabinetID,
		Name:        payload.Name,
		SIREN:       payload.SIREN,
		SIRET:       payload.SIRET,
		Phone:       payload.Phone,
		Email:       payload.Email,
		ContactName: payload.ContactName,
		Address:     payload.Address,
		Notes:       payload.Notes,
	}

	if err := r.clientRepo.Create(req.Context(), client); err != nil {
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

// updateClient handles PUT /api/v1/clients/{id}
func (r *Router) updateClient(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	client, err := r.clientRepo.GetByID(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get client")
		return
	}
	if client == nil {
		writeError(w, http.StatusNotFound, "Client not found")
		return
	}

	var payload UpdateClientRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	if payload.Name != nil {
		client.Name = *payload.Name
	}
	if payload.SIREN != nil {
		client.SIREN = payload.SIREN
	}
	if payload.SIRET != nil {
		client.SIRET = payload.SIRET
	}
	if payload.Phone != nil {
		client.Phone = payload.Phone
	}
	if payload.Email != nil {
		client.Email = payload.Email
	}
	if payload.ContactName != nil {
		client.ContactName = payload.ContactName
	}
	if payload.Address != nil {
		client.Address = payload.Address
	}
	if payload.Notes != nil {
		client.Notes = payload.Notes
	}
	if payload.WhatsAppOptedIn != nil {
		client.WhatsAppOptedIn = *payload.WhatsAppOptedIn
	}

	if err := r.clientRepo.Update(req.Context(), client); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update client")
		return
	}

	writeJSON(w, http.StatusOK, client)
}

// deleteClient handles DELETE /api/v1/clients/{id}
func (r *Router) deleteClient(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	if err := r.clientRepo.Delete(req.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete client")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
