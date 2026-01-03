package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
)

// listCampaigns handles GET /api/v1/campaigns
func (r *Router) listCampaigns(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.URL.Query().Get("cabinet_id")
	if cabinetIDStr == "" {
		// Default to demo cabinet for MVP
		cabinetIDStr = "00000000-0000-0000-0000-000000000001"
	}

	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	campaigns, err := r.campaignRepo.List(req.Context(), cabinetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}

	writeJSON(w, http.StatusOK, campaigns)
}

// createCampaign handles POST /api/v1/campaigns
func (r *Router) createCampaign(w http.ResponseWriter, req *http.Request) {
	var c models.Campaign
	if err := json.NewDecoder(req.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if c.CabinetID == uuid.Nil {
		// Default to demo cabinet
		c.CabinetID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	if err := r.campaignRepo.Create(req.Context(), &c); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

// getCampaign handles GET /api/v1/campaigns/{id}
func (r *Router) getCampaign(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	c, err := r.campaignRepo.GetByID(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}
	if c == nil {
		writeError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

// updateCampaign handles PATCH /api/v1/campaigns/{id}
func (r *Router) updateCampaign(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	c, err := r.campaignRepo.GetByID(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}
	if c == nil {
		writeError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	var update models.Campaign
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	c.Name = update.Name
	c.TriggerType = update.TriggerType
	c.IsActive = update.IsActive
	c.QuietHoursEnabled = update.QuietHoursEnabled
	if update.Steps != nil {
		c.Steps = update.Steps
	}

	if err := r.campaignRepo.Update(req.Context(), c); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

// deleteCampaign handles DELETE /api/v1/campaigns/{id}
func (r *Router) deleteCampaign(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	if err := r.campaignRepo.Delete(req.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
