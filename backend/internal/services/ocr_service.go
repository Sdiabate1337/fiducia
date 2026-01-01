package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// OCRService handles document OCR via GPT-4o Vision
type OCRService struct {
	apiKey      string
	httpClient  *http.Client
	storagePath string
}

// OCRResult contains extracted data from document
type OCRResult struct {
	RawText       string                 `json:"raw_text"`
	DocumentType  string                 `json:"document_type"` // invoice, receipt, bank_statement, other
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Confidence    float64                `json:"confidence"`
}

// ExtractedInvoiceData structured invoice/receipt data
type ExtractedInvoiceData struct {
	Date          string  `json:"date,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	Vendor        string  `json:"vendor,omitempty"`
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	Description   string  `json:"description,omitempty"`
}

// NewOCRService creates a new OCR service
func NewOCRService(apiKey, storagePath string) *OCRService {
	// Create storage directory
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		slog.Warn("failed to create OCR storage directory", "path", storagePath, "error", err)
	}

	return &OCRService{
		apiKey:      apiKey,
		storagePath: storagePath,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ProcessImage extracts text and data from an image
func (s *OCRService) ProcessImage(ctx context.Context, imagePath string) (*OCRResult, error) {
	// Read image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	// Determine media type
	mediaType := "image/jpeg"
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		mediaType = "image/png"
	case ".gif":
		mediaType = "image/gif"
	case ".webp":
		mediaType = "image/webp"
	}

	// Base64 encode
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	return s.processWithVision(ctx, base64Image, mediaType)
}

// ProcessImageFromURL extracts text from image at URL
func (s *OCRService) ProcessImageFromURL(ctx context.Context, imageURL string) (*OCRResult, error) {
	// Download image
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Determine media type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return s.processWithVision(ctx, base64Image, contentType)
}

// DownloadAndProcess downloads media from Twilio and processes it
func (s *OCRService) DownloadAndProcess(ctx context.Context, mediaURL, accountSID, authToken string) (*OCRResult, string, error) {
	// Download from Twilio (requires auth)
	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(accountSID, authToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download from Twilio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("Twilio download failed: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read media: %w", err)
	}

	// Determine file extension
	contentType := resp.Header.Get("Content-Type")
	ext := ".jpg"
	switch {
	case strings.Contains(contentType, "png"):
		ext = ".png"
	case strings.Contains(contentType, "pdf"):
		ext = ".pdf"
	case strings.Contains(contentType, "gif"):
		ext = ".gif"
	}

	// Save to storage
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.storagePath, filename)

	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to save file: %w", err)
	}

	// Process with OCR (skip PDFs for now, handle images)
	if ext == ".pdf" {
		return &OCRResult{
			RawText:      "PDF document - OCR not yet supported",
			DocumentType: "pdf",
			Confidence:   0,
		}, filePath, nil
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	result, err := s.processWithVision(ctx, base64Image, contentType)
	if err != nil {
		return nil, filePath, err
	}

	return result, filePath, nil
}

// processWithVision calls GPT-4o Vision API
func (s *OCRService) processWithVision(ctx context.Context, base64Image, mediaType string) (*OCRResult, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	prompt := `Analyze this document image and extract the following information in JSON format:

{
  "document_type": "invoice|receipt|bank_statement|other",
  "raw_text": "complete text content of the document",
  "date": "YYYY-MM-DD format if found",
  "amount": numeric value (without currency symbol),
  "currency": "EUR|USD|MAD|etc",
  "vendor": "company/merchant name",
  "invoice_number": "if found",
  "description": "brief description of what this document is about",
  "confidence": 0.0 to 1.0 indicating how confident you are
}

Be precise with amounts and dates. If a field is not found, omit it.`

	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:%s;base64,%s", mediaType, base64Image),
						},
					},
				},
			},
		},
		"max_tokens": 1000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := apiResp.Choices[0].Message.Content
	slog.Info("OCR response received", "content_length", len(content))

	// Extract JSON from response (may have markdown code blocks)
	jsonContent := extractJSON(content)

	// Parse extracted data
	var extractedData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &extractedData); err != nil {
		// If JSON parsing fails, return raw text
		return &OCRResult{
			RawText:      content,
			DocumentType: "unknown",
			Confidence:   0.5,
		}, nil
	}

	result := &OCRResult{
		DocumentType:  getString(extractedData, "document_type", "unknown"),
		RawText:       getString(extractedData, "raw_text", content),
		ExtractedData: extractedData,
		Confidence:    getFloat(extractedData, "confidence", 0.5),
	}

	return result, nil
}

// extractJSON extracts JSON from potential markdown code blocks
func extractJSON(content string) string {
	// Try to find JSON in code blocks
	re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}

	return content
}

func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
		}
	}
	return defaultVal
}
