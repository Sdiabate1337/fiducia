package voice

import (
	"bytes"
	"fmt"
	"os/exec"
)

// AudioConverter handles audio format conversions
type AudioConverter interface {
	ToOggOpus(input []byte) ([]byte, error)
	ConcatAudio(segments [][]byte) ([]byte, error)
}

// FFmpegConverter implements AudioConverter using FFmpeg
type FFmpegConverter struct {
	ffmpegPath string
}

// NewFFmpegConverter creates a new FFmpeg converter
func NewFFmpegConverter() *FFmpegConverter {
	// Try to find ffmpeg in PATH
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		path = "ffmpeg" // Fall back to default
	}
	return &FFmpegConverter{ffmpegPath: path}
}

// ToOggOpus converts audio to OGG/Opus format for WhatsApp voice notes
// WhatsApp requires OGG/Opus codec for native voice note playback
func (f *FFmpegConverter) ToOggOpus(input []byte) ([]byte, error) {
	// FFmpeg command:
	// ffmpeg -i pipe:0 -c:a libopus -b:a 32k -ar 48000 -ac 1 -application voip -f ogg pipe:1
	cmd := exec.Command(
		f.ffmpegPath,
		"-i", "pipe:0",           // Input from stdin
		"-c:a", "libopus",        // Opus codec
		"-b:a", "32k",            // 32kbps bitrate (good for voice)
		"-ar", "48000",           // 48kHz sample rate
		"-ac", "1",               // Mono
		"-application", "voip",   // Optimized for voice
		"-f", "ogg",              // OGG container
		"-y",                     // Overwrite output
		"pipe:1",                 // Output to stdout
	)

	cmd.Stdin = bytes.NewReader(input)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ConcatAudio concatenates multiple audio segments into one
// Used for combining cached phrases with generated variable parts
func (f *FFmpegConverter) ConcatAudio(segments [][]byte) ([]byte, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments to concatenate")
	}

	if len(segments) == 1 {
		return segments[0], nil
	}

	// Create a temporary concat file content
	// FFmpeg concat demuxer needs file paths, so we use filter_complex instead
	
	// Build filter for amix/concat
	var filterParts string
	for i := range segments {
		filterParts += fmt.Sprintf("[%d:a]", i)
	}
	filterParts += fmt.Sprintf("concat=n=%d:v=0:a=1[out]", len(segments))

	// Build input arguments
	args := make([]string, 0, len(segments)*2+6)
	for range segments {
		args = append(args, "-i", "pipe:0") // We'll handle this differently
	}
	args = append(args,
		"-filter_complex", filterParts,
		"-map", "[out]",
		"-c:a", "libopus",
		"-b:a", "32k",
		"-f", "ogg",
		"pipe:1",
	)

	// For multiple inputs, we need a different approach
	// Concatenate at the byte level after converting each to a common format
	var combined bytes.Buffer
	for i, segment := range segments {
		// Convert each segment to raw PCM first
		cmd := exec.Command(
			f.ffmpegPath,
			"-i", "pipe:0",
			"-f", "s16le",
			"-ar", "48000",
			"-ac", "1",
			"pipe:1",
		)
		cmd.Stdin = bytes.NewReader(segment)
		
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("ffmpeg PCM conversion failed for segment %d: %w", i, err)
		}
		
		combined.Write(stdout.Bytes())
	}

	// Convert combined PCM back to OGG/Opus
	cmd := exec.Command(
		f.ffmpegPath,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "1",
		"-i", "pipe:0",
		"-c:a", "libopus",
		"-b:a", "32k",
		"-f", "ogg",
		"pipe:1",
	)

	cmd.Stdin = &combined
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg final conversion failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
