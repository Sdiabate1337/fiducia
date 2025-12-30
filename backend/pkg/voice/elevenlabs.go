package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Service interface for voice operations
type Service interface {
	CloneVoice(name string, audioSample []byte) (voiceID string, err error)
	GenerateSpeech(voiceID, text string) (audioBytes []byte, err error)
	DeleteVoice(voiceID string) error
}

// ElevenLabsClient implements voice cloning via ElevenLabs
type ElevenLabsClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewElevenLabsClient creates a new ElevenLabs client
func NewElevenLabsClient(apiKey string) *ElevenLabsClient {
	return &ElevenLabsClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Voice cloning can take time
		},
		baseURL: "https://api.elevenlabs.io/v1",
	}
}

// CloneVoice creates a new voice clone from an audio sample
func (c *ElevenLabsClient) CloneVoice(name string, audioSample []byte) (string, error) {
	// Create multipart form data
	var body bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	
	// Name field
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"name\"\r\n\r\n")
	body.WriteString(name + "\r\n")
	
	// Description field
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"description\"\r\n\r\n")
	body.WriteString("Fiducia voice clone for cabinet collaborator\r\n")
	
	// Audio file
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"files\"; filename=\"sample.mp3\"\r\n")
	body.WriteString("Content-Type: audio/mpeg\r\n\r\n")
	body.Write(audioSample)
	body.WriteString("\r\n")
	
	body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	req, err := http.NewRequest("POST", c.baseURL+"/voices/add", &body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("elevenlabs error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response struct {
		VoiceID string `json:"voice_id"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.VoiceID, nil
}

// GenerateSpeech converts text to speech using the specified voice
func (c *ElevenLabsClient) GenerateSpeech(voiceID, text string) ([]byte, error) {
	payload := map[string]any{
		"text":     text,
		"model_id": "eleven_turbo_v2_5", // Fast French model
		"voice_settings": map[string]float64{
			"stability":        0.5,
			"similarity_boost": 0.8,
			"style":           0.0,
			"use_speaker_boost": true,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/text-to-speech/%s", c.baseURL, voiceID),
		bytes.NewReader(jsonPayload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elevenlabs error (status %d): %s", resp.StatusCode, string(errBody))
	}

	return io.ReadAll(resp.Body)
}

// DeleteVoice deletes a voice clone
func (c *ElevenLabsClient) DeleteVoice(voiceID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/voices/%s", c.baseURL, voiceID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elevenlabs error (status %d): %s", resp.StatusCode, string(errBody))
	}

	return nil
}
