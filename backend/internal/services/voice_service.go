package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/fiducia/backend/pkg/voice"
)

// VoiceService handles voice generation and storage
type VoiceService struct {
	elevenlabs  *voice.ElevenLabsClient
	ffmpeg      *voice.FFmpegConverter
	storagePath string
	baseURL     string
	cache       sync.Map // Cache for common phrases
}

// NewVoiceService creates a new voice service
func NewVoiceService(elevenLabsAPIKey, storagePath, baseURL string) *VoiceService {
	// Ensure storage directory exists
	if storagePath != "" {
		os.MkdirAll(storagePath, 0755)
	} else {
		storagePath = "/tmp/fiducia/voice"
		os.MkdirAll(storagePath, 0755)
	}

	return &VoiceService{
		elevenlabs:  voice.NewElevenLabsClient(elevenLabsAPIKey),
		ffmpeg:      voice.NewFFmpegConverter(),
		storagePath: storagePath,
		baseURL:     baseURL,
	}
}

// VoiceCloneRequest for cloning a collaborator's voice
type VoiceCloneRequest struct {
	CollaboratorID uuid.UUID
	Name           string
	AudioSample    []byte
}

// VoiceCloneResult contains the result of voice cloning
type VoiceCloneResult struct {
	VoiceID string
	Name    string
}

// CloneVoice creates a voice clone from an audio sample
func (s *VoiceService) CloneVoice(ctx context.Context, req VoiceCloneRequest) (*VoiceCloneResult, error) {
	voiceID, err := s.elevenlabs.CloneVoice(req.Name, req.AudioSample)
	if err != nil {
		return nil, fmt.Errorf("failed to clone voice: %w", err)
	}

	slog.Info("voice cloned successfully",
		"voice_id", voiceID,
		"collaborator_id", req.CollaboratorID,
		"name", req.Name,
	)

	return &VoiceCloneResult{
		VoiceID: voiceID,
		Name:    req.Name,
	}, nil
}

// DeleteVoice deletes a cloned voice
func (s *VoiceService) DeleteVoice(ctx context.Context, voiceID string) error {
	if err := s.elevenlabs.DeleteVoice(voiceID); err != nil {
		return fmt.Errorf("failed to delete voice: %w", err)
	}

	slog.Info("voice deleted", "voice_id", voiceID)
	return nil
}

// GenerateVoiceMessageRequest for generating a voice message
type GenerateVoiceMessageRequest struct {
	VoiceID       string
	Text          string
	PendingLineID uuid.UUID
	ConvertToOpus bool // Convert to OGG/Opus for WhatsApp
}

// GenerateVoiceMessageResult contains the generated audio
type GenerateVoiceMessageResult struct {
	AudioBytes []byte
	AudioURL   string
	Duration   float64
	Format     string
}

// GenerateVoiceMessage generates speech from text
func (s *VoiceService) GenerateVoiceMessage(ctx context.Context, req GenerateVoiceMessageRequest) (*GenerateVoiceMessageResult, error) {
	// Check cache for common phrases
	cacheKey := s.getCacheKey(req.VoiceID, req.Text)
	if cached, ok := s.cache.Load(cacheKey); ok {
		slog.Info("using cached voice message", "cache_key", cacheKey)
		return cached.(*GenerateVoiceMessageResult), nil
	}

	// Generate speech via ElevenLabs
	audioBytes, err := s.elevenlabs.GenerateSpeech(req.VoiceID, req.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate speech: %w", err)
	}

	// Convert to OGG/Opus for WhatsApp if requested
	format := "mp3"
	if req.ConvertToOpus {
		opusBytes, err := s.ffmpeg.ToOggOpus(audioBytes)
		if err != nil {
			slog.Warn("OGG/Opus conversion failed, using MP3", "error", err)
		} else {
			audioBytes = opusBytes
			format = "ogg"
		}
	}

	// Save to storage
	filename := fmt.Sprintf("%s_%s.%s", req.PendingLineID.String(), time.Now().Format("20060102_150405"), format)
	filepath := filepath.Join(s.storagePath, filename)

	if err := os.WriteFile(filepath, audioBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save audio file: %w", err)
	}

	// Generate URL
	audioURL := fmt.Sprintf("%s/audio/%s", s.baseURL, filename)

	result := &GenerateVoiceMessageResult{
		AudioBytes: audioBytes,
		AudioURL:   audioURL,
		Format:     format,
	}

	// Cache short common phrases
	if len(req.Text) < 100 {
		s.cache.Store(cacheKey, result)
	}

	slog.Info("voice message generated",
		"pending_line_id", req.PendingLineID,
		"format", format,
		"size_bytes", len(audioBytes),
		"audio_url", audioURL,
	)

	return result, nil
}

// getCacheKey generates a cache key for phrase caching
func (s *VoiceService) getCacheKey(voiceID, text string) string {
	h := sha256.New()
	h.Write([]byte(voiceID + "|" + text))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// GenerateRelanceVoice generates a voice relance message
func (s *VoiceService) GenerateRelanceVoice(ctx context.Context, voiceID string, clientName string, date string, amount string, label string, pendingLineID uuid.UUID) (*GenerateVoiceMessageResult, error) {
	// Script for voice message (shorter than text, more conversational)
	text := fmt.Sprintf(
		"Bonjour %s. Nous recherchons un justificatif pour l'opération du %s, d'un montant de %s euros, libellé %s. Merci de nous envoyer la pièce justificative. À bientôt.",
		clientName, date, amount, label,
	)

	return s.GenerateVoiceMessage(ctx, GenerateVoiceMessageRequest{
		VoiceID:       voiceID,
		Text:          text,
		PendingLineID: pendingLineID,
		ConvertToOpus: true, // Always convert to OGG/Opus for WhatsApp
	})
}

// CleanupOldAudio removes audio files older than specified duration
func (s *VoiceService) CleanupOldAudio(maxAge time.Duration) error {
	entries, err := os.ReadDir(s.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read storage dir: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	var deleted int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := filepath.Join(s.storagePath, entry.Name())
			if err := os.Remove(filepath); err == nil {
				deleted++
			}
		}
	}

	slog.Info("audio cleanup completed", "deleted", deleted)
	return nil
}

// GetAudioPath returns the full path to an audio file
func (s *VoiceService) GetAudioPath(filename string) string {
	return filepath.Join(s.storagePath, filename)
}
