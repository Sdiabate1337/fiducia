package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/fiducia/backend/internal/models"
)

// CSVImporter handles CSV file parsing and import
type CSVImporter struct {
	// Column mapping configuration
	dateFormats []string
}

// NewCSVImporter creates a new CSV importer
func NewCSVImporter() *CSVImporter {
	return &CSVImporter{
		dateFormats: []string{
			"02/01/2006", // DD/MM/YYYY (French)
			"2006-01-02", // YYYY-MM-DD (ISO)
			"02-01-2006", // DD-MM-YYYY
			"02.01.2006", // DD.MM.YYYY
			"2/1/2006",   // D/M/YYYY
			"01/02/2006", // MM/DD/YYYY (US)
		},
	}
}

// ImportResult contains the result of an import operation
type ImportResult struct {
	TotalRows    int                  `json:"total_rows"`
	ImportedRows int                  `json:"imported_rows"`
	FailedRows   int                  `json:"failed_rows"`
	Errors       []ImportError        `json:"errors,omitempty"`
	Lines        []models.PendingLine `json:"lines"`
}

// ImportError represents an error for a specific row
type ImportError struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
}

// ColumnMapping defines how CSV columns map to pending line fields
type ColumnMapping struct {
	AmountColumn  int `json:"amount_column"`
	DateColumn    int `json:"date_column"`
	LabelColumn   int `json:"label_column"`
	ClientColumn  int `json:"client_column,omitempty"`  // Optional
	AccountColumn int `json:"account_column,omitempty"` // Optional
}

// ClientColumnMapping defines how CSV columns map to client fields
type ClientColumnMapping struct {
	NameColumn  int `json:"name_column"`
	EmailColumn int `json:"email_column"`
	PhoneColumn int `json:"phone_column"`
	SiretColumn int `json:"siret_column"`
}

// ClientImportResult contains the result of a client import operation
type ClientImportResult struct {
	TotalRows    int             `json:"total_rows"`
	ImportedRows int             `json:"imported_rows"`
	FailedRows   int             `json:"failed_rows"`
	Errors       []ImportError   `json:"errors,omitempty"`
	Clients      []models.Client `json:"clients"`
}

// DetectedClientColumns represents auto-detected client column mappings
type DetectedClientColumns struct {
	Mapping    ClientColumnMapping `json:"mapping"`
	Confidence float64             `json:"confidence"`
	Headers    []string            `json:"headers"`
}

// DetectedColumns represents auto-detected column mappings
type DetectedColumns struct {
	Mapping    ColumnMapping `json:"mapping"`
	Confidence float64       `json:"confidence"`
	Headers    []string      `json:"headers"`
}

// ParseCSV parses a CSV file and returns pending lines
func (i *CSVImporter) ParseCSV(ctx context.Context, data []byte, cabinetID uuid.UUID, mapping *ColumnMapping) (*ImportResult, error) {
	// Detect and convert encoding
	data = i.ensureUTF8(data)

	// Detect delimiter
	delimiter := i.detectDelimiter(data)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have at least a header row and one data row")
	}

	// If no mapping provided, auto-detect
	if mapping == nil {
		detected := i.DetectColumns(records[0])
		mapping = &detected.Mapping
	}

	result := &ImportResult{
		TotalRows: len(records) - 1, // Exclude header
		Lines:     make([]models.PendingLine, 0, len(records)-1),
		Errors:    make([]ImportError, 0),
	}

	// Process data rows (skip header)
	for rowIdx, record := range records[1:] {
		line, err := i.parseRow(record, rowIdx+2, cabinetID, mapping)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowIdx + 2,
				Message: err.Error(),
			})
			result.FailedRows++
			continue
		}

		result.Lines = append(result.Lines, *line)
		result.ImportedRows++
	}

	return result, nil
}

// DetectColumns attempts to auto-detect column mappings from headers
func (i *CSVImporter) DetectColumns(headers []string) DetectedColumns {
	result := DetectedColumns{
		Headers: headers,
		Mapping: ColumnMapping{
			AmountColumn: -1,
			DateColumn:   -1,
			LabelColumn:  -1,
		},
		Confidence: 0,
	}

	amountPatterns := regexp.MustCompile(`(?i)(montant|amount|debit|credit|somme|total|ttc|ht)`)
	datePatterns := regexp.MustCompile(`(?i)(date|jour|day|operation|transaction|valeur)`)
	labelPatterns := regexp.MustCompile(`(?i)(libelle|libellé|label|description|designation|intitule|motif|reference)`)
	clientPatterns := regexp.MustCompile(`(?i)(client|tiers|fournisseur|raison.?sociale|nom|societe)`)
	accountPatterns := regexp.MustCompile(`(?i)(compte|account|numero|num)`)

	matches := 0
	for idx, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))

		if result.Mapping.AmountColumn == -1 && amountPatterns.MatchString(headerLower) {
			result.Mapping.AmountColumn = idx
			matches++
		} else if result.Mapping.DateColumn == -1 && datePatterns.MatchString(headerLower) {
			result.Mapping.DateColumn = idx
			matches++
		} else if result.Mapping.LabelColumn == -1 && labelPatterns.MatchString(headerLower) {
			result.Mapping.LabelColumn = idx
			matches++
		} else if result.Mapping.ClientColumn == 0 && clientPatterns.MatchString(headerLower) {
			result.Mapping.ClientColumn = idx
		} else if result.Mapping.AccountColumn == 0 && accountPatterns.MatchString(headerLower) {
			result.Mapping.AccountColumn = idx
		}
	}

	// Calculate confidence based on required fields found
	requiredFields := 3 // amount, date, label
	result.Confidence = float64(matches) / float64(requiredFields)

	// If auto-detect failed, use positional defaults for common ERP exports
	if result.Mapping.DateColumn == -1 && len(headers) > 0 {
		result.Mapping.DateColumn = 0 // First column often date
	}
	if result.Mapping.LabelColumn == -1 && len(headers) > 1 {
		result.Mapping.LabelColumn = 1 // Second often label
	}
	if result.Mapping.AmountColumn == -1 && len(headers) > 2 {
		result.Mapping.AmountColumn = len(headers) - 1 // Last often amount
	}

	return result
}

// PreviewCSV returns the first N rows for preview
func (i *CSVImporter) PreviewCSV(data []byte, maxRows int) ([][]string, error) {
	data = i.ensureUTF8(data)
	delimiter := i.detectDelimiter(data)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.LazyQuotes = true

	var rows [][]string
	for j := 0; j < maxRows+1; j++ { // +1 for header
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %w", j, err)
		}
		rows = append(rows, record)
	}

	return rows, nil
}

// parseRow parses a single CSV row into a PendingLine
func (i *CSVImporter) parseRow(record []string, rowNum int, cabinetID uuid.UUID, mapping *ColumnMapping) (*models.PendingLine, error) {
	if len(record) <= mapping.AmountColumn || len(record) <= mapping.DateColumn {
		return nil, fmt.Errorf("row has insufficient columns")
	}

	// Parse amount
	amountStr := strings.TrimSpace(record[mapping.AmountColumn])
	amount, err := i.parseAmount(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount '%s': %w", amountStr, err)
	}

	// Parse date
	dateStr := strings.TrimSpace(record[mapping.DateColumn])
	date, err := i.parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date '%s': %w", dateStr, err)
	}

	line := &models.PendingLine{
		ID:              uuid.New(),
		CabinetID:       cabinetID,
		Amount:          amount,
		TransactionDate: date,
		Status:          models.StatusPending,
		ContactCount:    0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Parse optional label
	if mapping.LabelColumn >= 0 && mapping.LabelColumn < len(record) {
		label := strings.TrimSpace(record[mapping.LabelColumn])
		if label != "" {
			line.BankLabel = &label
		}
	}

	// Parse optional account number
	if mapping.AccountColumn > 0 && mapping.AccountColumn < len(record) {
		account := strings.TrimSpace(record[mapping.AccountColumn])
		if account != "" {
			line.AccountNumber = &account
		}
	}

	return line, nil
}

// parseAmount parses a French or international format amount
func (i *CSVImporter) parseAmount(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, fmt.Errorf("empty amount")
	}

	// Remove currency symbols and spaces
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	// Handle French format (1 234,56) -> 1234.56
	if strings.Contains(s, ",") && !strings.Contains(s, ".") {
		// French format: comma as decimal separator
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") && strings.Contains(s, ".") {
		// Mixed format like 1.234,56 (European)
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}

	// Remove thousands separators (remaining spaces or dots before the decimal)
	parts := strings.Split(s, ".")
	if len(parts) == 2 && len(parts[1]) == 3 {
		// This looks like thousands separator, not decimal
		s = strings.ReplaceAll(s, ".", "")
	}

	return decimal.NewFromString(s)
}

// parseDate parses dates in various French and international formats
func (i *CSVImporter) parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	s = strings.TrimSpace(s)

	for _, format := range i.dateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format")
}

// ensureUTF8 converts the data to UTF-8 if necessary
func (i *CSVImporter) ensureUTF8(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}

	// Try to convert from ISO-8859-1 (Latin-1)
	var buf bytes.Buffer
	for _, b := range data {
		buf.WriteRune(rune(b))
	}
	return buf.Bytes()
}

// detectDelimiter detects the CSV delimiter from the first few lines
func (i *CSVImporter) detectDelimiter(data []byte) rune {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	delimiters := []rune{';', ',', '\t', '|'}
	counts := make(map[rune]int)

	// Check first 5 lines
	lines := 0
	for scanner.Scan() && lines < 5 {
		line := scanner.Text()
		for _, d := range delimiters {
			counts[d] += strings.Count(line, string(d))
		}
		lines++
	}

	// Return the most common delimiter
	maxCount := 0
	bestDelimiter := ';' // Default for French exports
	for d, count := range counts {
		if count > maxCount {
			maxCount = count
			bestDelimiter = d
		}
	}

	return bestDelimiter
}

// ValidateMapping checks if a mapping is valid for the given headers
func (i *CSVImporter) ValidateMapping(headers []string, mapping ColumnMapping) error {
	maxCol := len(headers) - 1

	if mapping.AmountColumn < 0 || mapping.AmountColumn > maxCol {
		return fmt.Errorf("amount column %d is out of range (0-%d)", mapping.AmountColumn, maxCol)
	}
	if mapping.DateColumn < 0 || mapping.DateColumn > maxCol {
		return fmt.Errorf("date column %d is out of range (0-%d)", mapping.DateColumn, maxCol)
	}
	if mapping.LabelColumn < 0 || mapping.LabelColumn > maxCol {
		return fmt.Errorf("label column %d is out of range (0-%d)", mapping.LabelColumn, maxCol)
	}

	return nil
}

// FormatAmount formats a decimal amount for display
func FormatAmount(d decimal.Decimal) string {
	// French format with space thousands separator and comma decimal
	s := d.StringFixed(2)

	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add thousands separators
	var result strings.Builder
	negative := false
	if intPart[0] == '-' {
		negative = true
		intPart = intPart[1:]
	}

	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(c)
	}

	formatted := result.String() + "," + decPart
	if negative {
		formatted = "-" + formatted
	}

	return formatted + " €"
}

// ParseClientsCSV parses a CSV file and returns clients
func (i *CSVImporter) ParseClientsCSV(ctx context.Context, data []byte, cabinetID uuid.UUID, mapping *ClientColumnMapping) (*ClientImportResult, error) {
	data = i.ensureUTF8(data)
	delimiter := i.detectDelimiter(data)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have at least a header row and one data row")
	}

	if mapping == nil {
		detected := i.DetectClientColumns(records[0])
		mapping = &detected.Mapping
	}

	result := &ClientImportResult{
		TotalRows: len(records) - 1,
		Clients:   make([]models.Client, 0, len(records)-1),
		Errors:    make([]ImportError, 0),
	}

	for rowIdx, record := range records[1:] {
		client, err := i.parseClientRow(record, rowIdx+2, cabinetID, mapping)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowIdx + 2,
				Message: err.Error(),
			})
			result.FailedRows++
			continue
		}

		result.Clients = append(result.Clients, *client)
		result.ImportedRows++
	}

	return result, nil
}

// DetectClientColumns attempts to auto-detect client column mappings
func (i *CSVImporter) DetectClientColumns(headers []string) DetectedClientColumns {
	result := DetectedClientColumns{
		Headers: headers,
		Mapping: ClientColumnMapping{
			NameColumn:  -1,
			EmailColumn: -1,
			PhoneColumn: -1,
			SiretColumn: -1,
		},
		Confidence: 0,
	}

	namePatterns := regexp.MustCompile(`(?i)(nom|name|raison.?sociale|societe|client|company)`)
	emailPatterns := regexp.MustCompile(`(?i)(email|mail|courriel)`)
	phonePatterns := regexp.MustCompile(`(?i)(tel|phone|mobile|portable|fixe)`)
	siretPatterns := regexp.MustCompile(`(?i)(siret|siren|tva|no.?vat)`)

	matches := 0
	for idx, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))

		if result.Mapping.NameColumn == -1 && namePatterns.MatchString(headerLower) {
			result.Mapping.NameColumn = idx
			matches++
		} else if result.Mapping.EmailColumn == -1 && emailPatterns.MatchString(headerLower) {
			result.Mapping.EmailColumn = idx
			matches++
		} else if result.Mapping.PhoneColumn == -1 && phonePatterns.MatchString(headerLower) {
			result.Mapping.PhoneColumn = idx
			matches++
		} else if result.Mapping.SiretColumn == -1 && siretPatterns.MatchString(headerLower) {
			result.Mapping.SiretColumn = idx
			matches++
		}
	}

	requiredFields := 1 // Name is required
	result.Confidence = float64(matches) / float64(requiredFields)
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	// Positional defaults
	if result.Mapping.NameColumn == -1 && len(headers) > 0 {
		result.Mapping.NameColumn = 0
	}
	if result.Mapping.EmailColumn == -1 && len(headers) > 1 {
		result.Mapping.EmailColumn = 1
	}

	return result
}

func (i *CSVImporter) parseClientRow(record []string, rowNum int, cabinetID uuid.UUID, mapping *ClientColumnMapping) (*models.Client, error) {
	if mapping.NameColumn < 0 || mapping.NameColumn >= len(record) {
		return nil, fmt.Errorf("missing name column")
	}

	name := strings.TrimSpace(record[mapping.NameColumn])
	if name == "" {
		return nil, fmt.Errorf("empty name")
	}

	client := &models.Client{
		ID:        uuid.New(),
		CabinetID: cabinetID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if mapping.EmailColumn >= 0 && mapping.EmailColumn < len(record) {
		email := strings.TrimSpace(record[mapping.EmailColumn])
		if email != "" {
			client.Email = &email
		}
	}

	if mapping.PhoneColumn >= 0 && mapping.PhoneColumn < len(record) {
		phone := strings.TrimSpace(record[mapping.PhoneColumn])
		// Basic phone cleaning (keep +, digits)
		if phone != "" {
			client.Phone = &phone
		}
	}

	if mapping.SiretColumn >= 0 && mapping.SiretColumn < len(record) {
		siret := strings.TrimSpace(record[mapping.SiretColumn])
		if siret != "" {
			client.SIRET = &siret
		}
	}

	return client, nil
}
