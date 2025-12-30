package whatsapp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client interface for WhatsApp operations
type Client interface {
	SendText(to, body string) (*MessageResponse, error)
	SendVoice(to, audioURL string) (*MessageResponse, error)
	SendTemplate(to, template string, params []string) (*MessageResponse, error)
	SendInteractive(to string, interactive InteractiveMessage) (*MessageResponse, error)
}

// MessageResponse represents the Twilio API response
type MessageResponse struct {
	MessageSID string `json:"sid"`
	Status     string `json:"status"`
}

// InteractiveMessage for buttons/lists
type InteractiveMessage struct {
	Type    string   `json:"type"`
	Body    string   `json:"body"`
	Buttons []Button `json:"buttons,omitempty"`
}

// Button for interactive messages
type Button struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// TwilioClient implements Client using Twilio API
type TwilioClient struct {
	accountSID  string
	authToken   string
	phoneNumber string
	baseURL     string
	httpClient  *http.Client
}

// NewTwilioClient creates a new Twilio WhatsApp client
func NewTwilioClient(accountSID, authToken, phoneNumber string) *TwilioClient {
	return &TwilioClient{
		accountSID:  accountSID,
		authToken:   authToken,
		phoneNumber: phoneNumber,
		baseURL:     fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID),
		httpClient:  &http.Client{},
	}
}

// SendText sends a text message via WhatsApp
func (c *TwilioClient) SendText(to, body string) (*MessageResponse, error) {
	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+c.phoneNumber)
	data.Set("Body", body)

	return c.makeRequest(data)
}

// SendVoice sends a voice message via WhatsApp
func (c *TwilioClient) SendVoice(to, audioURL string) (*MessageResponse, error) {
	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+c.phoneNumber)
	data.Set("MediaUrl", audioURL)

	return c.makeRequest(data)
}

// SendTemplate sends a pre-approved WhatsApp template message
func (c *TwilioClient) SendTemplate(to, templateName string, params []string) (*MessageResponse, error) {
	// Build ContentSid for template or use content variables
	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+c.phoneNumber)

	// For Twilio, templates are sent via Content API or as regular messages
	// This is a simplified implementation
	body := fmt.Sprintf("Template: %s, Params: %v", templateName, params)
	data.Set("Body", body)

	return c.makeRequest(data)
}

// SendInteractive sends an interactive message with buttons
func (c *TwilioClient) SendInteractive(to string, interactive InteractiveMessage) (*MessageResponse, error) {
	// Twilio uses different approach for interactive messages
	// This is a placeholder implementation
	data := url.Values{}
	data.Set("To", "whatsapp:"+to)
	data.Set("From", "whatsapp:"+c.phoneNumber)
	data.Set("Body", interactive.Body)

	return c.makeRequest(data)
}

// makeRequest makes an HTTP request to Twilio
func (c *TwilioClient) makeRequest(data url.Values) (*MessageResponse, error) {
	req, err := http.NewRequest("POST", c.baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(c.accountSID+":"+c.authToken),
	))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Twilio API error: status %d", resp.StatusCode)
	}

	// For now, return a mock response
	// In production, parse the actual Twilio response
	return &MessageResponse{
		MessageSID: "SM" + fmt.Sprintf("%032d", 0), // Placeholder
		Status:     "queued",
	}, nil
}

// MockClient is a mock implementation for testing
type MockClient struct{}

// NewMockClient creates a mock WhatsApp client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// SendText mock implementation
func (c *MockClient) SendText(to, body string) (*MessageResponse, error) {
	return &MessageResponse{
		MessageSID: "MOCK_" + to,
		Status:     "sent",
	}, nil
}

// SendVoice mock implementation
func (c *MockClient) SendVoice(to, audioURL string) (*MessageResponse, error) {
	return &MessageResponse{
		MessageSID: "MOCK_VOICE_" + to,
		Status:     "sent",
	}, nil
}

// SendTemplate mock implementation
func (c *MockClient) SendTemplate(to, template string, params []string) (*MessageResponse, error) {
	return &MessageResponse{
		MessageSID: "MOCK_TEMPLATE_" + to,
		Status:     "sent",
	}, nil
}

// SendInteractive mock implementation
func (c *MockClient) SendInteractive(to string, interactive InteractiveMessage) (*MessageResponse, error) {
	return &MessageResponse{
		MessageSID: "MOCK_INTERACTIVE_" + to,
		Status:     "sent",
	}, nil
}
