package ocr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// Service interface for OCR operations
type Service interface {
	ExtractFromImage(imageData []byte) (*Result, error)
	ExtractFromURL(imageURL string) (*Result, error)
}

// Result represents extracted data from an image
type Result struct {
	Amount     *decimal.Decimal `json:"amount,omitempty"`
	Date       *string          `json:"date,omitempty"` // Format: YYYY-MM-DD
	Merchant   *string          `json:"merchant,omitempty"`
	RawText    string           `json:"raw_text,omitempty"`
	Confidence float64          `json:"confidence"`
}

// GPT4VisionClient implements OCR via GPT-4o-mini Vision
type GPT4VisionClient struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewGPT4VisionClient creates a new GPT-4 Vision client
func NewGPT4VisionClient(apiKey string) *GPT4VisionClient {
	return &GPT4VisionClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gpt-4o-mini", // Cost-effective vision model
	}
}

// OCR extraction prompt optimized for French receipts/invoices
const ocrPrompt = `Tu es un expert OCR spécialisé dans l'extraction de données de tickets de caisse et factures.

Analyse cette image et extrait les informations suivantes :
- montant_ttc : le montant total TTC (nombre décimal, ex: 42.50)
- date : la date de la transaction (format YYYY-MM-DD)
- marchand : le nom du commerce ou fournisseur

RÈGLES IMPORTANTES :
1. Si une information est illisible ou absente, mets null
2. Pour le montant, cherche "TOTAL", "TTC", "À PAYER" ou le montant le plus grand
3. Pour la date, convertis toujours au format YYYY-MM-DD
4. Réponds UNIQUEMENT en JSON valide, pas d'explication

Format de réponse EXACTE :
{"montant_ttc": 42.50, "date": "2024-01-15", "marchand": "Carrefour", "confiance": 0.95}`

// ExtractFromImage extracts data from image bytes
func (c *GPT4VisionClient) ExtractFromImage(imageData []byte) (*Result, error) {
	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	// Detect image type (simplified)
	imageType := "image/jpeg"
	if len(imageData) > 8 && string(imageData[0:8]) == "\x89PNG\r\n\x1a\n" {
		imageType = "image/png"
	}

	return c.callVisionAPI(fmt.Sprintf("data:%s;base64,%s", imageType, base64Image))
}

// ExtractFromURL extracts data from an image URL
func (c *GPT4VisionClient) ExtractFromURL(imageURL string) (*Result, error) {
	return c.callVisionAPI(imageURL)
}

func (c *GPT4VisionClient) callVisionAPI(imageSource string) (*Result, error) {
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": ocrPrompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url":    imageSource,
							"detail": "high", // High detail for OCR
						},
					},
				},
			},
		},
		"max_tokens": 500,
		"temperature": 0.1, // Low temperature for consistent extraction
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(jsonPayload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := openAIResp.Choices[0].Message.Content

	// Parse extracted JSON
	var extracted struct {
		MontantTTC *float64 `json:"montant_ttc"`
		Date       *string  `json:"date"`
		Marchand   *string  `json:"marchand"`
		Confiance  float64  `json:"confiance"`
	}
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		// If JSON parsing fails, return with low confidence
		return &Result{
			RawText:    content,
			Confidence: 0.1,
		}, nil
	}

	result := &Result{
		Date:       extracted.Date,
		Merchant:   extracted.Marchand,
		RawText:    content,
		Confidence: extracted.Confiance,
	}

	if extracted.MontantTTC != nil {
		amount := decimal.NewFromFloat(*extracted.MontantTTC)
		result.Amount = &amount
	}

	return result, nil
}
