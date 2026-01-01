package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/database"
	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
	"github.com/fiducia/backend/internal/services"
	"github.com/fiducia/backend/pkg/whatsapp"
)

// Router holds dependencies for HTTP handlers
type Router struct {
	db          *database.DB
	cfg         *config.Config
	mux         *http.ServeMux
	importer    *services.CSVImporter
	lineRepo    *repository.PendingLineRepository
	waClient    *whatsapp.TwilioClient
	voiceSvc    *services.VoiceService
	ocrSvc      *services.OCRService
	matchingSvc *services.MatchingService
	docRepo     *repository.DocumentRepository
	clientRepo  *repository.ClientRepository
	msgRepo     *repository.MessageRepository
	voiceRepo   *repository.VoiceSettingsRepository
}

// NewRouter creates a new HTTP router with all routes
func NewRouter(db *database.DB, cfg *config.Config) *Router {
	lineRepo := repository.NewPendingLineRepository(db.Pool)
	docRepo := repository.NewDocumentRepository(db.Pool)
	clientRepo := repository.NewClientRepository(db.Pool)
	msgRepo := repository.NewMessageRepository(db.Pool)
	ocrSvc := services.NewOCRService(cfg.OpenAIAPIKey, "/tmp/fiducia/documents")
	matchingSvc := services.NewMatchingService(docRepo, lineRepo)

	r := &Router{
		db:          db,
		cfg:         cfg,
		mux:         http.NewServeMux(),
		importer:    services.NewCSVImporter(),
		lineRepo:    lineRepo,
		waClient:    whatsapp.NewTwilioClient(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioPhoneNumber),
		voiceSvc:    services.NewVoiceService(cfg.ElevenLabsAPIKey, "/tmp/fiducia/voice", cfg.BaseURL),
		ocrSvc:      ocrSvc,
		matchingSvc: matchingSvc,
		docRepo:     docRepo,
		clientRepo:  clientRepo,
		msgRepo:     msgRepo,
		voiceRepo:   repository.NewVoiceSettingsRepository(db.Pool),
	}

	r.registerRoutes()
	return r
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// GetPool returns the database pool
func (r *Router) GetPool() *pgxpool.Pool {
	return r.db.Pool
}

// registerRoutes sets up all API routes
func (r *Router) registerRoutes() {
	// Health check
	r.mux.HandleFunc("GET /health", r.healthCheck)
	r.mux.HandleFunc("GET /api/v1/health", r.healthCheck)

	// Cabinets
	r.mux.HandleFunc("GET /api/v1/cabinets", r.listCabinets)
	r.mux.HandleFunc("POST /api/v1/cabinets", r.createCabinet)
	r.mux.HandleFunc("GET /api/v1/cabinets/{id}", r.getCabinet)

	// Clients
	r.mux.HandleFunc("GET /api/v1/cabinets/{cabinet_id}/clients", r.listClients)
	r.mux.HandleFunc("POST /api/v1/cabinets/{cabinet_id}/clients", r.createClient)
	r.mux.HandleFunc("GET /api/v1/clients/{id}", r.getClient)

	// Pending Lines (471)
	r.mux.HandleFunc("GET /api/v1/cabinets/{cabinet_id}/pending-lines", r.listPendingLines)
	r.mux.HandleFunc("GET /api/v1/cabinets/{cabinet_id}/pending-lines/stats", r.getPendingLinesStats)
	r.mux.HandleFunc("POST /api/v1/cabinets/{cabinet_id}/pending-lines", r.createPendingLine)
	r.mux.HandleFunc("GET /api/v1/pending-lines/{id}", r.getPendingLine)
	r.mux.HandleFunc("PATCH /api/v1/pending-lines/{id}", r.updatePendingLine)

	// Import - REAL IMPLEMENTATIONS
	r.mux.HandleFunc("POST /api/v1/cabinets/{cabinet_id}/import/preview", r.previewCSV)
	r.mux.HandleFunc("POST /api/v1/cabinets/{cabinet_id}/import/csv", r.importCSV)
	r.mux.HandleFunc("GET /api/v1/import/{id}/status", r.getImportStatus)

	// Messages
	r.mux.HandleFunc("GET /api/v1/pending-lines/{id}/messages", r.listMessages)
	r.mux.HandleFunc("POST /api/v1/pending-lines/{id}/messages", r.sendMessage)

	// Webhooks (both with and without api prefix for Twilio convenience)
	r.mux.HandleFunc("POST /api/v1/webhook/whatsapp", r.whatsappWebhook)
	r.mux.HandleFunc("POST /webhook/whatsapp", r.whatsappWebhook)

	// Documents
	r.mux.HandleFunc("GET /api/v1/pending-lines/{id}/documents", r.listDocuments)
	r.mux.HandleFunc("POST /api/v1/documents/{id}/approve", r.approveDocument)
	r.mux.HandleFunc("POST /api/v1/documents/{id}/reject", r.rejectDocument)

	// Matching Proposals
	r.mux.HandleFunc("GET /api/v1/cabinets/{cabinet_id}/proposals", r.listProposals)
	r.mux.HandleFunc("POST /api/v1/proposals/{id}/approve", r.approveProposal)
	r.mux.HandleFunc("POST /api/v1/proposals/{id}/reject", r.rejectProposal)

	// Exports
	r.mux.HandleFunc("POST /api/v1/cabinets/{cabinet_id}/exports", r.createExport)
	r.mux.HandleFunc("GET /api/v1/exports/{id}", r.getExport)

	// Voice (Sprint 3)
	r.mux.HandleFunc("POST /api/v1/collaborators/{id}/voice/clone", r.cloneVoice)
	r.mux.HandleFunc("DELETE /api/v1/collaborators/{id}/voice", r.deleteVoice)
	r.mux.HandleFunc("POST /api/v1/voice/generate", r.generateVoice)
	r.mux.HandleFunc("GET /audio/{filename}", r.serveAudio)
	r.mux.HandleFunc("GET /api/v1/documents/content/{filename}", r.serveDocument)
}

// ============================================
// HEALTH CHECK
// ============================================

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "fiducia-api",
	}
	writeJSON(w, http.StatusOK, response)
}

// ============================================
// CABINET HANDLERS
// ============================================

func (r *Router) listCabinets(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"cabinets": []any{}, "total": 0})
}

func (r *Router) createCabinet(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Cabinet created"})
}

func (r *Router) getCabinet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

// ============================================
// CLIENT HANDLERS
// ============================================

func (r *Router) listClients(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": []any{}, "total": 0})
}

func (r *Router) createClient(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Client created"})
}

func (r *Router) getClient(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

// ============================================
// PENDING LINES HANDLERS
// ============================================

func (r *Router) listPendingLines(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	filter := repository.PendingLineFilter{
		CabinetID: cabinetID,
		Limit:     50,
	}

	// Parse query params
	if status := req.URL.Query().Get("status"); status != "" {
		// TODO: Add status filter
	}

	result, err := r.lineRepo.List(req.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list pending lines")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) getPendingLinesStats(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	stats, err := r.lineRepo.GetStats(req.Context(), cabinetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (r *Router) createPendingLine(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Pending line created"})
}

func (r *Router) getPendingLine(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	line, err := r.lineRepo.GetByID(req.Context(), id)
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

func (r *Router) updatePendingLine(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "message": "Updated"})
}

// ============================================
// IMPORT HANDLERS - REAL IMPLEMENTATION
// ============================================

func (r *Router) previewCSV(w http.ResponseWriter, req *http.Request) {
	// Parse multipart form
	if err := req.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Get preview rows (max 10)
	maxRows := 10
	if rowsParam := req.URL.Query().Get("rows"); rowsParam != "" {
		if n, err := strconv.Atoi(rowsParam); err == nil && n > 0 && n <= 50 {
			maxRows = n
		}
	}

	rows, err := r.importer.PreviewCSV(data, maxRows)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse CSV: "+err.Error())
		return
	}

	// Detect columns from headers
	detected := services.DetectedColumns{Confidence: 0}
	if len(rows) > 0 {
		detected = r.importer.DetectColumns(rows[0])
	}

	response := map[string]any{
		"filename":   header.Filename,
		"size":       header.Size,
		"rows":       rows,
		"total_rows": len(rows) - 1, // Exclude header
		"detected":   detected,
	}

	writeJSON(w, http.StatusOK, response)
}

func (r *Router) importCSV(w http.ResponseWriter, req *http.Request) {
	cabinetIDStr := req.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Parse multipart form
	if err := req.ParseMultipartForm(50 << 20); err != nil { // 50MB max
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read file")
		return
	}

	// Parse optional mapping from form field
	var mapping *services.ColumnMapping
	if mappingJSON := req.FormValue("mapping"); mappingJSON != "" {
		if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid mapping JSON")
			return
		}
	}

	// Parse CSV
	result, err := r.importer.ParseCSV(req.Context(), data, cabinetID, mapping)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse CSV: "+err.Error())
		return
	}

	// Set source file on all lines
	for i := range result.Lines {
		result.Lines[i].SourceFile = &header.Filename
		rowNum := i + 2 // 1-indexed, skip header
		result.Lines[i].SourceRowNumber = &rowNum
	}

	// Insert lines in batch
	if len(result.Lines) > 0 {
		if err := r.lineRepo.CreateBatch(req.Context(), result.Lines); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save pending lines: "+err.Error())
			return
		}
	}

	response := map[string]any{
		"batch_id":      uuid.New().String(),
		"total_rows":    result.TotalRows,
		"imported_rows": result.ImportedRows,
		"failed_rows":   result.FailedRows,
		"errors":        result.Errors,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (r *Router) getImportStatus(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "completed"})
}

// ============================================
// MESSAGE HANDLERS
// ============================================

func (r *Router) listMessages(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	msgRepo := repository.NewMessageRepository(r.db.Pool)
	messages, err := msgRepo.ListByPendingLine(req.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list messages")
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"total":    len(messages),
	})
}

func (r *Router) sendMessage(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	pendingLineID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	// Parse request body
	var body struct {
		MessageType   string `json:"message_type"`
		CustomMessage string `json:"custom_message"`
		Immediate     bool   `json:"immediate"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		body.MessageType = "text"
	}
	if body.MessageType == "" {
		body.MessageType = "text"
	}

	// Get pending line with client
	line, err := r.lineRepo.GetByID(req.Context(), pendingLineID)
	if err != nil || line == nil {
		writeError(w, http.StatusNotFound, "Pending line not found")
		return
	}

	// Check if client is assigned
	if line.ClientID == nil {
		writeError(w, http.StatusBadRequest, "No client assigned to this pending line")
		return
	}

	// Get client
	clientRepo := repository.NewClientRepository(r.db.Pool)
	client, err := clientRepo.GetByID(req.Context(), *line.ClientID)
	if err != nil || client == nil {
		writeError(w, http.StatusBadRequest, "Client not found")
		return
	}

	if client.Phone == nil || *client.Phone == "" {
		writeError(w, http.StatusBadRequest, "Client has no phone number")
		return
	}

	// Generate message content
	content := body.CustomMessage
	if content == "" {
		amount := line.Amount.StringFixed(2)
		date := line.TransactionDate.Format("02/01/2006")
		label := "une opération"
		if line.BankLabel != nil {
			label = *line.BankLabel
		}
		content = fmt.Sprintf(
			"Bonjour %s,\n\nNous recherchons un justificatif pour l'opération suivante :\n\n📅 Date : %s\n💰 Montant : %s €\n📝 Libellé : %s\n\nMerci de nous envoyer la pièce justificative.\n\nCordialement,\nVotre cabinet comptable",
			client.Name, date, amount, label,
		)
	}

	// Create message record
	msgRepo := repository.NewMessageRepository(r.db.Pool)
	msg := &models.Message{
		ID:            uuid.New(),
		PendingLineID: &pendingLineID,
		ClientID:      line.ClientID,
		Direction:     models.DirectionOutbound,
		MessageType:   models.MessageType(body.MessageType),
		Content:       &content,
		Status:        models.MsgStatusQueued,
	}

	if err := msgRepo.Create(req.Context(), msg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create message: "+err.Error())
		return
	}

	// Send via Twilio if immediate mode or in production
	var waMessageID string
	var audioURL string
	if body.Immediate && r.waClient != nil && r.cfg.TwilioAccountSID != "" {
		// Update status to sending
		msgRepo.UpdateStatus(req.Context(), msg.ID, models.MsgStatusSending, nil)

		var resp *whatsapp.MessageResponse
		var err error

		// Handle voice messages
		if body.MessageType == "voice" && r.voiceSvc != nil && r.cfg.ElevenLabsAPIKey != "" {
			// Generate voice message
			amount := line.Amount.StringFixed(2)
			date := line.TransactionDate.Format("02/01/2006")
			label := "une opération"
			if line.BankLabel != nil {
				label = *line.BankLabel
			}

			// Determine Voice ID (use cloned voice if available)
			voiceID := r.cfg.ElevenLabsVoiceID // Default

			// For this MVP, we use the test collaborator ID used in frontend
			testCollaboratorID, _ := uuid.Parse("22222222-2222-2222-2222-222222222222")
			if voiceSetting, _ := r.voiceRepo.GetByCollaboratorID(req.Context(), testCollaboratorID); voiceSetting != nil {
				voiceID = voiceSetting.VoiceID
				slog.Info("using cloned voice", "voice_id", voiceID, "name", voiceSetting.Name)
			}

			voiceResult, voiceErr := r.voiceSvc.GenerateRelanceVoice(
				req.Context(),
				voiceID, // Use determined voice ID
				client.Name,
				date,
				amount,
				label,
				pendingLineID,
			)
			if voiceErr != nil {
				errMsg := "Voice generation failed: " + voiceErr.Error()
				msgRepo.SetError(req.Context(), msg.ID, errMsg)
				writeJSON(w, http.StatusCreated, map[string]any{
					"message": "Génération vocale échouée",
					"id":      msg.ID.String(),
					"status":  "failed",
					"error":   errMsg,
				})
				return
			}

			audioURL = voiceResult.AudioURL
			// In development with Twilio Sandbox, MediaUrl doesn't work well
			// Send text message instead, but keep the generated audio for future use
			if r.cfg.IsDevelopment() {
				// Fallback to text message in sandbox mode
				resp, err = r.waClient.SendText(*client.Phone, content+" (🎙️ Audio: "+audioURL+")")
			} else {
				// In production, send voice note via Twilio
				resp, err = r.waClient.SendVoice(*client.Phone, audioURL)
			}
		} else {
			// Send text via Twilio
			resp, err = r.waClient.SendText(*client.Phone, content)
		}

		if err != nil {
			// Mark as failed but don't return error - still log the message
			errMsg := err.Error()
			msgRepo.SetError(req.Context(), msg.ID, errMsg)
			writeJSON(w, http.StatusCreated, map[string]any{
				"message": "Envoi échoué",
				"id":      msg.ID.String(),
				"status":  "failed",
				"error":   errMsg,
			})
			return
		}

		// Update with WhatsApp message ID
		waMessageID = resp.MessageSID
		msgRepo.UpdateStatus(req.Context(), msg.ID, models.MsgStatusSent, &waMessageID)
		msg.Status = models.MsgStatusSent
	}

	// Update pending line status (only if not already in later state)
	if line.Status == models.StatusPending {
		line.Status = models.StatusContacted
	}
	line.ContactCount++
	now := time.Now()
	line.LastContactedAt = &now
	r.lineRepo.Update(req.Context(), line)

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":       "Relance " + string(msg.Status),
		"id":            msg.ID.String(),
		"status":        msg.Status,
		"wa_message_id": waMessageID,
		"audio_url":     audioURL,
		"content":       content,
	})
}

func (r *Router) whatsappWebhook(w http.ResponseWriter, req *http.Request) {
	// Parse form data
	if err := req.ParseForm(); err != nil {
		slog.Error("failed to parse webhook form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract Twilio payload
	messageSid := req.FormValue("MessageSid")
	from := strings.TrimPrefix(req.FormValue("From"), "whatsapp:")
	body := req.FormValue("Body")
	numMedia := req.FormValue("NumMedia")
	mediaUrl0 := req.FormValue("MediaUrl0")
	mediaType0 := req.FormValue("MediaContentType0")

	slog.Info("received WhatsApp webhook",
		"message_sid", messageSid,
		"from", from,
		"body_length", len(body),
		"num_media", numMedia,
	)

	// Find client by phone
	client, _ := r.clientRepo.GetByPhoneGlobal(req.Context(), from)
	var clientID *uuid.UUID
	if client != nil {
		clientID = &client.ID
	}

	// Save incoming message
	waID := messageSid
	msg := &models.Message{
		ClientID:    clientID,
		Direction:   models.DirectionInbound,
		MessageType: models.TypeText,
		Content:     &body,
		Status:      models.MsgStatusDelivered,
		WAMessageID: &waID,
	}

	if numMedia != "0" && numMedia != "" {
		msg.MessageType = models.TypeMedia
		msg.MediaURL = &mediaUrl0
	}

	if err := r.msgRepo.Create(req.Context(), msg); err != nil {
		slog.Error("failed to save incoming message", "error", err)
	}

	// Process media if present (use background context for async processing)
	if numMedia != "0" && numMedia != "" && mediaUrl0 != "" {
		go r.processIncomingMedia(context.Background(), mediaUrl0, mediaType0, clientID, msg.ID)
	}

	// Respond with TwiML
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Response></Response>`)
}

// processIncomingMedia handles OCR processing in background
func (r *Router) processIncomingMedia(ctx context.Context, mediaURL, mediaType string, clientID *uuid.UUID, messageID uuid.UUID) {
	slog.Info("processing incoming media", "url", mediaURL, "type", mediaType)

	// Download and process with OCR
	ocrResult, filePath, err := r.ocrSvc.DownloadAndProcess(
		ctx,
		mediaURL,
		r.cfg.TwilioAccountSID,
		r.cfg.TwilioAuthToken,
	)

	// Create document record
	doc := &repository.Document{
		ClientID:        clientID,
		MessageID:       &messageID,
		FilePath:        filePath,
		FileType:        &mediaType,
		TwilioMediaURL:  &mediaURL,
		MatchConfidence: decimal.Zero,
		MatchStatus:     "pending",
	}

	if err != nil {
		slog.Error("OCR processing failed", "error", err)
		doc.OCRStatus = "failed"
		errStr := err.Error()
		doc.OCRError = &errStr
	} else {
		doc.OCRStatus = "completed"
		doc.OCRText = &ocrResult.RawText
		doc.OCRData = ocrResult.ExtractedData
		slog.Info("OCR completed", "text_length", len(ocrResult.RawText), "doc_type", ocrResult.DocumentType)
	}

	if err := r.docRepo.Create(ctx, doc); err != nil {
		slog.Error("failed to save document", "error", err)
		return
	}

	slog.Info("document saved", "doc_id", doc.ID, "ocr_status", doc.OCRStatus)

	// Auto-match if OCR succeeded
	if doc.OCRStatus == "completed" && clientID != nil {
		proposal, err := r.matchingSvc.AutoMatch(ctx, doc)
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

// ============================================
// DOCUMENT HANDLERS
// ============================================

func (r *Router) listDocuments(w http.ResponseWriter, req *http.Request) {
	pendingLineID := req.PathValue("id")
	id, err := uuid.Parse(pendingLineID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid pending line ID")
		return
	}

	docs, err := r.docRepo.GetByPendingLine(req.Context(), id)
	if err != nil {
		slog.Error("failed to list documents", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}

	// Transform to API response
	result := make([]map[string]any, 0, len(docs))
	for _, doc := range docs {
		item := map[string]any{
			"id":               doc.ID.String(),
			"file_path":        doc.FilePath,
			"file_type":        doc.FileType,
			"ocr_status":       doc.OCRStatus,
			"ocr_text":         doc.OCRText,
			"ocr_data":         doc.OCRData,
			"match_confidence": doc.MatchConfidence,
			"match_status":     doc.MatchStatus,
			"created_at":       doc.CreatedAt,
		}
		result = append(result, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"documents": result,
		"total":     len(result),
	})
}

func (r *Router) approveDocument(w http.ResponseWriter, req *http.Request) {
	docID := req.PathValue("id")
	id, err := uuid.Parse(docID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	// Get document
	doc, err := r.docRepo.GetByID(req.Context(), id)
	if err != nil || doc == nil {
		writeError(w, http.StatusNotFound, "Document not found")
		return
	}

	// Approve the match
	if err := r.docRepo.ApproveMatch(req.Context(), id, nil); err != nil {
		slog.Error("failed to approve document", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to approve")
		return
	}

	// Update pending line status if linked
	if doc.PendingLineID != nil {
		line, _ := r.lineRepo.GetByID(req.Context(), *doc.PendingLineID)
		if line != nil {
			line.Status = models.StatusValidated
			r.lineRepo.Update(req.Context(), line)
		}
	}

	slog.Info("document approved", "doc_id", id, "pending_line_id", doc.PendingLineID)

	writeJSON(w, http.StatusOK, map[string]string{
		"id":     docID,
		"status": "approved",
	})
}

func (r *Router) rejectDocument(w http.ResponseWriter, req *http.Request) {
	docID := req.PathValue("id")
	id, err := uuid.Parse(docID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	// Get document
	doc, err := r.docRepo.GetByID(req.Context(), id)
	if err != nil || doc == nil {
		writeError(w, http.StatusNotFound, "Document not found")
		return
	}

	// Reject the match
	if err := r.docRepo.RejectMatch(req.Context(), id, nil); err != nil {
		slog.Error("failed to reject document", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to reject")
		return
	}

	slog.Info("document rejected", "doc_id", id)

	writeJSON(w, http.StatusOK, map[string]string{
		"id":     docID,
		"status": "rejected",
	})
}

// ============================================
// PROPOSAL HANDLERS
// ============================================

func (r *Router) listProposals(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"proposals": []any{}, "total": 0})
}

func (r *Router) approveProposal(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "approved"})
}

func (r *Router) rejectProposal(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "rejected"})
}

// ============================================
// EXPORT HANDLERS
// ============================================

func (r *Router) createExport(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Export started"})
}

func (r *Router) getExport(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": "ready"})
}

// ============================================
// VOICE HANDLERS (Sprint 3)
// ============================================

func (r *Router) cloneVoice(w http.ResponseWriter, req *http.Request) {
	collaboratorIDStr := req.PathValue("id")
	collaboratorID, err := uuid.Parse(collaboratorIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid collaborator ID")
		return
	}

	// Parse multipart form (max 10MB audio file)
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	file, header, err := req.FormFile("audio")
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

	name := req.FormValue("name")
	if name == "" {
		name = header.Filename
	}

	result, err := r.voiceSvc.CloneVoice(req.Context(), services.VoiceCloneRequest{
		CollaboratorID: collaboratorID,
		Name:           name,
		AudioSample:    audioBytes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to clone voice: "+err.Error())
		return
	}

	// Save to database
	setting := &repository.VoiceSetting{
		CollaboratorID: collaboratorID,
		VoiceID:        result.VoiceID,
		Name:           result.Name,
	}
	if err := r.voiceRepo.Create(req.Context(), setting); err != nil {
		slog.Error("failed to save voice setting", "error", err)
		// Don't fail request, just log error
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"voice_id": result.VoiceID,
		"name":     result.Name,
		"message":  "Voice cloned successfully",
	})
}

func (r *Router) deleteVoice(w http.ResponseWriter, req *http.Request) {
	voiceID := req.URL.Query().Get("voice_id")
	if voiceID == "" {
		writeError(w, http.StatusBadRequest, "voice_id is required")
		return
	}

	if err := r.voiceSvc.DeleteVoice(req.Context(), voiceID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete voice: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Voice deleted"})
}

func (r *Router) generateVoice(w http.ResponseWriter, req *http.Request) {
	var body struct {
		VoiceID string `json:"voice_id"`
		Text    string `json:"text"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if body.VoiceID == "" || body.Text == "" {
		writeError(w, http.StatusBadRequest, "voice_id and text are required")
		return
	}

	result, err := r.voiceSvc.GenerateVoiceMessage(req.Context(), services.GenerateVoiceMessageRequest{
		VoiceID:       body.VoiceID,
		Text:          body.Text,
		PendingLineID: uuid.New(),
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

func (r *Router) serveAudio(w http.ResponseWriter, req *http.Request) {
	filename := req.PathValue("filename")
	if filename == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Prevent directory traversal
	filename = filepath.Base(filename)
	audioPath := r.voiceSvc.GetAudioPath(filename)

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

	http.ServeFile(w, req, audioPath)
}

func (r *Router) serveDocument(w http.ResponseWriter, req *http.Request) {
	filename := req.PathValue("filename")
	if filename == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Prevent directory traversal
	filename = filepath.Base(filename)
	docPath := filepath.Join("/tmp/fiducia/documents", filename)

	// Check if file exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Set content type based on extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, req, docPath)
}

// ============================================
// HELPERS
// ============================================

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
