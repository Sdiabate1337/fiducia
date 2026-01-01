package services

import (
	"context"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

// MatchingService handles document-to-pending-line matching
type MatchingService struct {
	docRepo  *repository.DocumentRepository
	lineRepo *repository.PendingLineRepository
}

// MatchProposal represents a potential match
type MatchProposal struct {
	DocumentID    uuid.UUID       `json:"document_id"`
	PendingLineID uuid.UUID       `json:"pending_line_id"`
	Confidence    decimal.Decimal `json:"confidence"`
	Reasons       []string        `json:"reasons"`
}

// NewMatchingService creates a new matching service
func NewMatchingService(docRepo *repository.DocumentRepository, lineRepo *repository.PendingLineRepository) *MatchingService {
	return &MatchingService{
		docRepo:  docRepo,
		lineRepo: lineRepo,
	}
}

// FindMatches finds potential matches for a document
func (s *MatchingService) FindMatches(ctx context.Context, doc *repository.Document) ([]MatchProposal, error) {
	// Get client's pending lines
	if doc.ClientID == nil {
		slog.Warn("document has no client ID, cannot match")
		return nil, nil
	}

	// Get all pending lines for this client's cabinet
	// For simplicity, we'll search all pending lines and filter by client
	lines, err := s.lineRepo.ListByClient(ctx, *doc.ClientID)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		slog.Info("no pending lines found for client", "client_id", doc.ClientID)
		return nil, nil
	}

	// Extract document data
	docAmount := extractAmount(doc.OCRData)
	docDate := extractDate(doc.OCRData)
	docVendor := extractVendor(doc.OCRData)

	slog.Info("matching document",
		"doc_id", doc.ID,
		"amount", docAmount,
		"date", docDate,
		"vendor", docVendor,
	)

	var proposals []MatchProposal

	for _, line := range lines {
		if line.Status == models.StatusValidated {
			continue // Skip already validated lines
		}

		confidence, reasons := s.calculateMatchScore(line, docAmount, docDate, docVendor)

		if confidence >= 0.3 { // Minimum threshold for proposal
			proposals = append(proposals, MatchProposal{
				DocumentID:    doc.ID,
				PendingLineID: line.ID,
				Confidence:    decimal.NewFromFloat(confidence),
				Reasons:       reasons,
			})
		}
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(proposals)-1; i++ {
		for j := i + 1; j < len(proposals); j++ {
			if proposals[j].Confidence.GreaterThan(proposals[i].Confidence) {
				proposals[i], proposals[j] = proposals[j], proposals[i]
			}
		}
	}

	return proposals, nil
}

// AutoMatch attempts to automatically match a document to a pending line
func (s *MatchingService) AutoMatch(ctx context.Context, doc *repository.Document) (*MatchProposal, error) {
	proposals, err := s.FindMatches(ctx, doc)
	if err != nil {
		return nil, err
	}

	if len(proposals) == 0 {
		return nil, nil
	}

	// Check if best match has high enough confidence for auto-match
	best := proposals[0]
	if best.Confidence.GreaterThanOrEqual(decimal.NewFromFloat(0.85)) {
		// Auto-match with high confidence
		if err := s.docRepo.UpdateMatch(ctx, doc.ID, &best.PendingLineID, best.Confidence, "auto_matched"); err != nil {
			return nil, err
		}

		// Update pending line status
		line, _ := s.lineRepo.GetByID(ctx, best.PendingLineID)
		if line != nil {
			line.Status = models.StatusReceived
			s.lineRepo.Update(ctx, line)
		}

		slog.Info("auto-matched document",
			"doc_id", doc.ID,
			"line_id", best.PendingLineID,
			"confidence", best.Confidence,
		)

		return &best, nil
	}

	// Propose match for manual review
	if err := s.docRepo.UpdateMatch(ctx, doc.ID, &best.PendingLineID, best.Confidence, "pending"); err != nil {
		return nil, err
	}

	return &best, nil
}

// calculateMatchScore calculates match score between document and pending line
func (s *MatchingService) calculateMatchScore(line *models.PendingLine, docAmount float64, docDate, docVendor string) (float64, []string) {
	var score float64
	var reasons []string

	// Amount matching (most important)
	lineAmount, _ := line.Amount.Float64()
	if docAmount > 0 && lineAmount > 0 {
		amountDiff := math.Abs(docAmount - lineAmount)
		if amountDiff < 0.01 {
			score += 0.5
			reasons = append(reasons, "Montant exact")
		} else if amountDiff < 1.0 {
			score += 0.4
			reasons = append(reasons, "Montant proche (±1€)")
		} else if amountDiff < 5.0 {
			score += 0.2
			reasons = append(reasons, "Montant similaire (±5€)")
		}
	}

	// Date matching
	if docDate != "" {
		docTime, err := parseFlexibleDate(docDate)
		if err == nil {
			dateDiff := line.TransactionDate.Sub(docTime).Hours() / 24
			if math.Abs(dateDiff) < 1 {
				score += 0.3
				reasons = append(reasons, "Date exacte")
			} else if math.Abs(dateDiff) < 7 {
				score += 0.2
				reasons = append(reasons, "Date proche (±7j)")
			} else if math.Abs(dateDiff) < 30 {
				score += 0.1
				reasons = append(reasons, "Date similaire (±30j)")
			}
		}
	}

	// Vendor/Label matching
	if docVendor != "" && line.BankLabel != nil {
		vendorLower := strings.ToLower(docVendor)
		labelLower := strings.ToLower(*line.BankLabel)

		if strings.Contains(labelLower, vendorLower) || strings.Contains(vendorLower, labelLower) {
			score += 0.2
			reasons = append(reasons, "Fournisseur correspondant")
		} else {
			// Check for common abbreviations
			vendorParts := strings.Fields(vendorLower)
			for _, part := range vendorParts {
				if len(part) > 3 && strings.Contains(labelLower, part) {
					score += 0.1
					reasons = append(reasons, "Fournisseur partiellement correspondant")
					break
				}
			}
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score, reasons
}

// extractAmount extracts amount from OCR data
func extractAmount(data map[string]interface{}) float64 {
	if data == nil {
		return 0
	}

	if v, ok := data["amount"]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case string:
			// Remove currency symbols and parse
			cleaned := strings.ReplaceAll(t, "€", "")
			cleaned = strings.ReplaceAll(cleaned, "$", "")
			cleaned = strings.ReplaceAll(cleaned, ",", ".")
			cleaned = strings.TrimSpace(cleaned)
			if f, err := strconv.ParseFloat(cleaned, 64); err == nil {
				return f
			}
		}
	}

	return 0
}

// extractDate extracts date from OCR data
func extractDate(data map[string]interface{}) string {
	if data == nil {
		return ""
	}

	if v, ok := data["date"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

// extractVendor extracts vendor from OCR data
func extractVendor(data map[string]interface{}) string {
	if data == nil {
		return ""
	}

	if v, ok := data["vendor"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

// parseFlexibleDate parses various date formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
		"01/02/2006",
		"2006/01/02",
		"02 Jan 2006",
		"02 January 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}
