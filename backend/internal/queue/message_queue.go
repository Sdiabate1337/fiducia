package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageJob represents a WhatsApp message to send
type MessageJob struct {
	ID             string    `json:"id"`
	PendingLineID  string    `json:"pending_line_id"`
	ClientID       string    `json:"client_id"`
	Phone          string    `json:"phone"`
	MessageType    string    `json:"message_type"` // text, voice, template
	Content        string    `json:"content"`
	TemplateName   string    `json:"template_name,omitempty"`
	TemplateParams []string  `json:"template_params,omitempty"`
	AudioURL       string    `json:"audio_url,omitempty"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	Attempts       int       `json:"attempts"`
	CreatedAt      time.Time `json:"created_at"`
}

// MessageQueue handles message queueing with anti-ban logic
type MessageQueue struct {
	client        *redis.Client
	queueKey      string
	delayedKey    string
	processingKey string
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(redisURL string) (*MessageQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("connected to redis")

	return &MessageQueue{
		client:        client,
		queueKey:      "fiducia:messages:queue",
		delayedKey:    "fiducia:messages:delayed",
		processingKey: "fiducia:messages:processing",
	}, nil
}

// Close closes the Redis connection
func (q *MessageQueue) Close() error {
	return q.client.Close()
}

// Anti-ban configuration
const (
	MinJitterSeconds  = 30   // Minimum delay between messages
	MaxJitterSeconds  = 180  // Maximum delay between messages
	MaxMessagesPerDay = 3    // Max messages per phone per day
	TypingDelayMs     = 2000 // Simulated typing delay
)

// Enqueue adds a message to the queue with anti-ban jittering
func (q *MessageQueue) Enqueue(ctx context.Context, job *MessageJob) error {
	// Check rate limit for this phone
	if limited, err := q.isRateLimited(ctx, job.Phone); err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	} else if limited {
		return fmt.Errorf("rate limit exceeded for phone %s", job.Phone)
	}

	// Calculate jittered delay
	jitter := MinJitterSeconds + rand.Intn(MaxJitterSeconds-MinJitterSeconds)
	job.ScheduledAt = time.Now().Add(time.Duration(jitter) * time.Second)
	job.CreatedAt = time.Now()

	// Serialize job
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to delayed sorted set (score = scheduled timestamp)
	score := float64(job.ScheduledAt.Unix())
	if err := q.client.ZAdd(ctx, q.delayedKey, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Increment rate limit counter
	if err := q.incrementRateLimit(ctx, job.Phone); err != nil {
		slog.Warn("failed to increment rate limit", "error", err)
	}

	slog.Info("message enqueued",
		"job_id", job.ID,
		"phone", job.Phone,
		"scheduled_at", job.ScheduledAt,
		"jitter_seconds", jitter,
	)

	return nil
}

// EnqueueImmediate adds a message for immediate sending (for testing)
func (q *MessageQueue) EnqueueImmediate(ctx context.Context, job *MessageJob) error {
	job.ScheduledAt = time.Now()
	job.CreatedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := q.client.RPush(ctx, q.queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves the next ready message from the queue
func (q *MessageQueue) Dequeue(ctx context.Context) (*MessageJob, error) {
	// First, move any ready delayed jobs to the main queue
	if err := q.moveReadyJobs(ctx); err != nil {
		slog.Warn("failed to move ready jobs", "error", err)
	}

	// Pop from main queue
	result, err := q.client.BLPop(ctx, 1*time.Second, q.queueKey).Result()
	if err == redis.Nil {
		return nil, nil // No jobs available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	var job MessageJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Add to processing set
	q.client.SAdd(ctx, q.processingKey, result[1])

	return &job, nil
}

// Complete marks a job as completed
func (q *MessageQueue) Complete(ctx context.Context, job *MessageJob) error {
	data, _ := json.Marshal(job)
	return q.client.SRem(ctx, q.processingKey, data).Err()
}

// Retry puts a failed job back in the queue with exponential backoff
func (q *MessageQueue) Retry(ctx context.Context, job *MessageJob) error {
	job.Attempts++

	// Exponential backoff: 1min, 2min, 4min, 8min, 16min
	backoff := time.Duration(1<<job.Attempts) * time.Minute
	if backoff > 30*time.Minute {
		backoff = 30 * time.Minute
	}

	job.ScheduledAt = time.Now().Add(backoff)

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Remove from processing
	q.client.SRem(ctx, q.processingKey, data)

	// Add back to delayed queue
	score := float64(job.ScheduledAt.Unix())
	return q.client.ZAdd(ctx, q.delayedKey, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

// moveReadyJobs moves jobs that are ready to be sent from delayed to main queue
func (q *MessageQueue) moveReadyJobs(ctx context.Context) error {
	now := float64(time.Now().Unix())

	// Get all jobs with score <= now
	jobs, err := q.client.ZRangeByScore(ctx, q.delayedKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	if err != nil {
		return err
	}

	for _, jobData := range jobs {
		// Move to main queue
		if err := q.client.RPush(ctx, q.queueKey, jobData).Err(); err != nil {
			continue
		}
		// Remove from delayed
		q.client.ZRem(ctx, q.delayedKey, jobData)
	}

	return nil
}

// isRateLimited checks if a phone has exceeded the daily message limit
func (q *MessageQueue) isRateLimited(ctx context.Context, phone string) (bool, error) {
	key := fmt.Sprintf("fiducia:ratelimit:%s:%s", phone, time.Now().Format("2006-01-02"))
	count, err := q.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return count >= MaxMessagesPerDay, nil
}

// incrementRateLimit increments the daily message count for a phone
func (q *MessageQueue) incrementRateLimit(ctx context.Context, phone string) error {
	key := fmt.Sprintf("fiducia:ratelimit:%s:%s", phone, time.Now().Format("2006-01-02"))
	pipe := q.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetQueueStats returns queue statistics
func (q *MessageQueue) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	pending, _ := q.client.LLen(ctx, q.queueKey).Result()
	delayed, _ := q.client.ZCard(ctx, q.delayedKey).Result()
	processing, _ := q.client.SCard(ctx, q.processingKey).Result()

	return map[string]int64{
		"pending":    pending,
		"delayed":    delayed,
		"processing": processing,
	}, nil
}
