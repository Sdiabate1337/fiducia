package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client interface for WhatsApp operations
type Client interface {
	SendText(to, message string) (*MessageResponse, error)
	SendVoice(to, audioURL string) (*MessageResponse, error)
	SendTemplate(to, templateName string, params []string) (*MessageResponse, error)
	SendInteractive(to string, body string, buttons []Button) (*MessageResponse, error)
}

// Button represents an interactive button
type Button struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// MessageResponse represents a successful message response
type MessageResponse struct {
	MessageSID    string `json:"message_sid"`
	Status        string `json:"status"`
	DateCreated   string `json:"date_created"`
}

// TwilioClient implements WhatsApp via Twilio
type TwilioClient struct {
	accountSID  string
	authToken   string
	fromNumber  string
	httpClient  *http.Client
	baseURL     string
}

// NewTwilioClient creates a new Twilio WhatsApp client
func NewTwilioClient(accountSID, authToken, fromNumber string) *TwilioClient {
	return &TwilioClient{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s", accountSID),
	}
}

// SendText sends a text message via WhatsApp
func (c *TwilioClient) SendText(to, message string) (*MessageResponse, error) {
	return c.sendMessage(to, map[string]string{
		"Body": message,
	})
}

// SendVoice sends a voice note via WhatsApp
func (c *TwilioClient) SendVoice(to, audioURL string) (*MessageResponse, error) {
	return c.sendMessage(to, map[string]string{
		"MediaUrl": audioURL,
	})
}

// SendTemplate sends a template message via WhatsApp
func (c *TwilioClient) SendTemplate(to, templateName string, params []string) (*MessageResponse, error) {
	// Build content variables JSON
	contentVars := make(map[string]string)
	for i, param := range params {
		contentVars[fmt.Sprintf("%d", i+1)] = param
	}
	
	varsJSON, err := json.Marshal(contentVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template params: %w", err)
	}

	return c.sendMessage(to, map[string]string{
		"ContentSid":       templateName,
		"ContentVariables": string(varsJSON),
	})
}

// SendInteractive sends an interactive message with buttons
func (c *TwilioClient) SendInteractive(to string, body string, buttons []Button) (*MessageResponse, error) {
	// Build interactive message JSON
	persistent := make([]map[string]string, len(buttons))
	for i, btn := range buttons {
		persistent[i] = map[string]string{
			"type":  "reply",
			"reply": fmt.Sprintf(`{"id":"%s","title":"%s"}`, btn.ID, btn.Title),
		}
	}

	interactiveJSON, err := json.Marshal(map[string]any{
		"type": "button",
		"body": map[string]string{
			"text": body,
		},
		"action": map[string]any{
			"buttons": persistent,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interactive message: %w", err)
	}

	return c.sendMessage(to, map[string]string{
		"Body":           body,
		"PersistentAction": string(interactiveJSON),
	})
}

// sendMessage is the internal method to send messages
func (c *TwilioClient) sendMessage(to string, params map[string]string) (*MessageResponse, error) {
	formData := fmt.Sprintf("From=whatsapp:%s&To=whatsapp:%s", c.fromNumber, to)
	for key, value := range params {
		formData += fmt.Sprintf("&%s=%s", key, value)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/Messages.json", c.baseURL),
		bytes.NewBufferString(formData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("twilio error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Sid         string `json:"sid"`
		Status      string `json:"status"`
		DateCreated string `json:"date_created"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &MessageResponse{
		MessageSID:  response.Sid,
		Status:      response.Status,
		DateCreated: response.DateCreated,
	}, nil
}
