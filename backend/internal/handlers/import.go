package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
	"github.com/fiducia/backend/internal/services"
)

// ImportHandler handles CSV import requests
type ImportHandler struct {
	importer   *services.CSVImporter
	lineRepo   *repository.PendingLineRepository
	batchRepo  *repository.ImportBatchRepository
	clientRepo *repository.ClientRepository
}

// NewImportHandler creates a new import handler
func NewImportHandler(
	lineRepo *repository.PendingLineRepository,
	batchRepo *repository.ImportBatchRepository,
	clientRepo *repository.ClientRepository,
) *ImportHandler {
	return &ImportHandler{
		importer:   services.NewCSVImporter(),
		lineRepo:   lineRepo,
		batchRepo:  batchRepo,
		clientRepo: clientRepo,
	}
}

// PreviewCSV handles POST /api/v1/cabinets/{cabinet_id}/import/preview
func (h *ImportHandler) PreviewCSV(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
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
	if rowsParam := r.URL.Query().Get("rows"); rowsParam != "" {
		if n, err := strconv.Atoi(rowsParam); err == nil && n > 0 && n <= 50 {
			maxRows = n
		}
	}

	rows, err := h.importer.PreviewCSV(data, maxRows)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse CSV: "+err.Error())
		return
	}

	// Detect columns from headers
	detected := services.DetectedColumns{Confidence: 0}
	if len(rows) > 0 {
		detected = h.importer.DetectColumns(rows[0])
	}

	response := map[string]any{
		"filename":    header.Filename,
		"size":        header.Size,
		"rows":        rows,
		"total_rows":  len(rows) - 1, // Exclude header
		"detected":    detected,
	}

	writeJSON(w, http.StatusOK, response)
}

// ImportCSVRequest represents the import request body
type ImportCSVRequest struct {
	Mapping *services.ColumnMapping `json:"mapping,omitempty"`
}

// ImportCSV handles POST /api/v1/cabinets/{cabinet_id}/import/csv
func (h *ImportHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max
		writeError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
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
	if mappingJSON := r.FormValue("mapping"); mappingJSON != "" {
		if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid mapping JSON")
			return
		}
	}

	// Create import batch record
	batch := &models.ImportBatch{
		ID:        uuid.New(),
		CabinetID: cabinetID,
		Filename:  &header.Filename,
		Status:    "processing",
	}
	fileType := "csv"
	batch.FileType = &fileType

	if err := h.batchRepo.Create(r.Context(), batch); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create import batch")
		return
	}

	// Parse CSV
	result, err := h.importer.ParseCSV(r.Context(), data, cabinetID, mapping)
	if err != nil {
		h.batchRepo.UpdateStatus(r.Context(), batch.ID, "failed", 0, 0, map[string]any{
			"error": err.Error(),
		})
		writeError(w, http.StatusBadRequest, "Failed to parse CSV: "+err.Error())
		return
	}

	// Set import batch ID on all lines
	for i := range result.Lines {
		result.Lines[i].ImportBatchID = &batch.ID
		result.Lines[i].SourceFile = &header.Filename
		rowNum := i + 2 // 1-indexed, skip header
		result.Lines[i].SourceRowNumber = &rowNum
	}

	// Insert lines in batch
	if len(result.Lines) > 0 {
		if err := h.lineRepo.CreateBatch(r.Context(), result.Lines); err != nil {
			h.batchRepo.UpdateStatus(r.Context(), batch.ID, "failed", 0, 0, map[string]any{
				"error": err.Error(),
			})
			writeError(w, http.StatusInternalServerError, "Failed to save pending lines")
			return
		}
	}

	// Convert errors to map for storage
	var errorsMap map[string]any
	if len(result.Errors) > 0 {
		errorsMap = map[string]any{
			"rows": result.Errors,
		}
	}

	// Update batch status
	h.batchRepo.UpdateStatus(r.Context(), batch.ID, "completed",
		result.ImportedRows, result.FailedRows, errorsMap)

	response := map[string]any{
		"batch_id":      batch.ID,
		"total_rows":    result.TotalRows,
		"imported_rows": result.ImportedRows,
		"failed_rows":   result.FailedRows,
		"errors":        result.Errors,
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetImportStatus handles GET /api/v1/import/{id}/status
func (h *ImportHandler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid import ID")
		return
	}

	batch, err := h.batchRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get import batch")
		return
	}
	if batch == nil {
		writeError(w, http.StatusNotFound, "Import batch not found")
		return
	}

	writeJSON(w, http.StatusOK, batch)
}

// ListImports handles GET /api/v1/cabinets/{cabinet_id}/imports
func (h *ImportHandler) ListImports(w http.ResponseWriter, r *http.Request) {
	cabinetIDStr := r.PathValue("cabinet_id")
	cabinetID, err := uuid.Parse(cabinetIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cabinet ID")
		return
	}

	limit := 20
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if n, err := strconv.Atoi(limitParam); err == nil {
			limit = n
		}
	}

	batches, err := h.batchRepo.List(r.Context(), cabinetID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list imports")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"imports": batches,
		"total":   len(batches),
	})
}
