package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fiducia/backend/internal/middleware"
	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
	"github.com/google/uuid"
)

// listCabinets handles GET /api/v1/cabinets
// For admin purposes? Or just redirect to getCabinet(Me)?
// For now, let's keep it simple: strict access control
func (r *Router) listCabinets(w http.ResponseWriter, req *http.Request) {
	// Usually super-admin only.
	// For "Me" context, we should probably just return the user's cabinet.
	cabinetID, ok := middleware.GetCabinetID(req.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Just return the single cabinet in a list for now, or implement real list if needed
	repo := repository.NewCabinetRepository(r.db)
	cab, err := repo.GetByID(req.Context(), cabinetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch cabinet")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"cabinets": []models.Cabinet{*cab},
		"total":    1,
	})
}

// getCabinet handles GET /api/v1/cabinets/{id}
func (r *Router) getCabinet(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Security check: ensure user belongs to this cabinet
	userCabID, ok := middleware.GetCabinetID(req.Context())
	if !ok || userCabID != id {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	repo := repository.NewCabinetRepository(r.db)
	cab, err := repo.GetByID(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get cabinet")
		return
	}
	if cab == nil {
		writeError(w, http.StatusNotFound, "Cabinet not found")
		return
	}

	writeJSON(w, http.StatusOK, cab)
}

// createCabinet handles POST /api/v1/cabinets
// NOTE: Creation is actually done via Auth/Register now.
// This endpoint might be deprecated or admin-only.
func (r *Router) createCabinet(w http.ResponseWriter, req *http.Request) {
	// ... logic if needed, otherwise:
	writeError(w, http.StatusMethodNotAllowed, "Use /auth/register to create a cabinet")
}

// updateCabinet handles PATCH /api/v1/cabinets/{id}
func (r *Router) updateCabinet(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Security check
	userCabID, ok := middleware.GetCabinetID(req.Context())
	if !ok || userCabID != id {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	var body struct {
		Name                *string `json:"name"`
		SIRET               *string `json:"siret"`
		Address             *string `json:"address"`
		Phone               *string `json:"phone"`
		Email               *string `json:"email"`
		OnboardingCompleted *bool   `json:"onboarding_completed"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	repo := repository.NewCabinetRepository(r.db)
	cab, err := repo.GetByID(req.Context(), id)
	if err != nil || cab == nil {
		writeError(w, http.StatusNotFound, "Cabinet not found")
		return
	}

	// Update fields
	if body.Name != nil {
		cab.Name = *body.Name
	}
	if body.SIRET != nil {
		cab.SIRET = body.SIRET
	}
	if body.Address != nil {
		cab.Address = body.Address
	}
	if body.Phone != nil {
		cab.Phone = body.Phone
	}
	if body.Email != nil {
		cab.Email = body.Email
	}
	if body.OnboardingCompleted != nil {
		cab.OnboardingCompleted = *body.OnboardingCompleted
	}

	if err := repo.Update(req.Context(), cab); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update cabinet")
		return
	}

	writeJSON(w, http.StatusOK, cab)
}

// getOnboardingStatus handles GET /api/v1/cabinets/{id}/onboarding-status
func (r *Router) getOnboardingStatus(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Security check
	userCabID, ok := middleware.GetCabinetID(req.Context())
	if !ok || userCabID != id {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	userID, ok := middleware.GetUserID(req.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 1. Check Clients
	clientRepo := repository.NewClientRepository(r.db.Pool)
	clientCount, err := clientRepo.Count(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count clients")
		return
	}

	// 2. Check Pending Lines
	lineRepo := repository.NewPendingLineRepository(r.db.Pool)
	lineStats, err := lineRepo.GetStats(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get line stats")
		return
	}
	lineCount := 0
	if total, ok := lineStats["total"].(int); ok {
		lineCount = total
	}

	// 3. Check Voice (for current user)
	voiceRepo := repository.NewVoiceSettingsRepository(r.db.Pool)
	voiceSettings, err := voiceRepo.GetByCollaboratorID(req.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check voice settings")
		return
	}
	voiceConfigured := voiceSettings != nil

	status := map[string]any{
		"has_clients":      clientCount > 0,
		"client_count":     clientCount,
		"has_lines":        lineCount > 0,
		"line_count":       lineCount,
		"voice_configured": voiceConfigured,
	}

	writeJSON(w, http.StatusOK, status)
}
