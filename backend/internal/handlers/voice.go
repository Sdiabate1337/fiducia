package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/services"
)

// VoiceHandler handles voice-related API requests
type VoiceHandler struct {
	voiceSvc *services.VoiceService
	pool     *pgxpool.Pool
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(voiceSvc *services.VoiceService, pool *pgxpool.Pool) *VoiceHandler {
	return &VoiceHandler{
		voiceSvc: voiceSvc,
		pool:     pool,
	}
}

// CloneVoice handles POST /api/v1/collaborators/{id}/voice/clone
func (h *VoiceHandler) CloneVoice(w http.ResponseWriter, r *http.Request) {
	collaboratorIDStr := r.PathValue("id")
	collaboratorID, err := uuid.Parse(collaboratorIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid collaborator ID")
		return
	}

	// Parse multipart form (max 10MB audio file)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	// Get audio file
	file, header, err := r.FormFile("audio")
	if err != nil {
		writeError(w, http.StatusBadRequest, "No audio file provided")
		return
	}
	defer file.Close()

	audioBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read audio file")
		return
	}

	// Get name from form
	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}

	// Clone voice
	result, err := h.voiceSvc.CloneVoice(r.Context(), services.VoiceCloneRequest{
		CollaboratorID: collaboratorID,
		Name:           name,
		AudioSample:    audioBytes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to clone voice: "+err.Error())
		return
	}

	// TODO: Store voice_id in collaborators table
	// UPDATE collaborators SET voice_id = $1 WHERE id = $2

	writeJSON(w, http.StatusCreated, map[string]any{
		"voice_id": result.VoiceID,
		"name":     result.Name,
		"message":  "Voice cloned successfully",
	})
}

// DeleteVoice handles DELETE /api/v1/collaborators/{id}/voice
func (h *VoiceHandler) DeleteVoice(w http.ResponseWriter, r *http.Request) {
	voiceID := r.URL.Query().Get("voice_id")
	if voiceID == "" {
		writeError(w, http.StatusBadRequest, "voice_id is required")
		return
	}

	if err := h.voiceSvc.DeleteVoice(r.Context(), voiceID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete voice: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Voice deleted"})
}

// GenerateVoiceRequest for generating a voice message
type GenerateVoiceRequest struct {
	VoiceID string `json:"voice_id"`
	Text    string `json:"text"`
}

// GenerateVoice handles POST /api/v1/voice/generate
func (h *VoiceHandler) GenerateVoice(w http.ResponseWriter, r *http.Request) {
	var req GenerateVoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.VoiceID == "" || req.Text == "" {
		writeError(w, http.StatusBadRequest, "voice_id and text are required")
		return
	}

	result, err := h.voiceSvc.GenerateVoiceMessage(r.Context(), services.GenerateVoiceMessageRequest{
		VoiceID:       req.VoiceID,
		Text:          req.Text,
		PendingLineID: uuid.New(), // Generate a temp ID for standalone generation
		ConvertToOpus: true,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate voice: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"audio_url": result.AudioURL,
		"format":    result.Format,
		"size":      len(result.AudioBytes),
	})
}

// ServeAudio handles GET /audio/{filename}
func (h *VoiceHandler) ServeAudio(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Prevent directory traversal
	filename = filepath.Base(filename)
	audioPath := h.voiceSvc.GetAudioPath(filename)

	// Check if file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Set content type based on extension
	ext := filepath.Ext(filename)
	switch ext {
	case ".ogg":
		w.Header().Set("Content-Type", "audio/ogg")
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, r, audioPath)
}
