package migrations

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// TokenBucket implements rate limiting for AI API calls
type TokenBucket struct {
	tokens     float64
	capacity   float64
	rate       float64 // tokens per second
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket for rate limiting
func NewTokenBucket(capacity, rate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastUpdate: time.Now(),
	}
}

// Consume attempts to consume the specified number of tokens
// Returns true if successful, false if not enough tokens available
func (tb *TokenBucket) Consume(tokens float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	// Refill tokens based on elapsed time
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	// Check if we have enough tokens
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

// Global token bucket for AI rate limiting
// 15,000 tokens/minute = 250 tokens/second
var aiTokenBucket *TokenBucket

// StartAIQueueProcessor starts the background AI queue processor
func StartAIQueueProcessor(app core.App) {
	log.Println("🚀 Starting AI Queue Processor...")

	// Initialize token bucket from environment
	rateLimitTokens := getEnvFloat("AI_RATE_LIMIT_TOKENS", 15000)
	rateLimitWindow := getEnvFloat("AI_RATE_LIMIT_WINDOW", 60) // seconds
	ratePerSecond := rateLimitTokens / rateLimitWindow

	aiTokenBucket = NewTokenBucket(rateLimitTokens, ratePerSecond)

	log.Printf("✅ Token bucket initialized: %.0f tokens, %.2f tokens/sec", rateLimitTokens, ratePerSecond)

	// Get queue processing interval from environment
	intervalSec := getEnvFloat("QUEUE_PROCESS_INTERVAL", 5)
	interval := time.Duration(intervalSec) * time.Second

	// Start the background processor
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			processPendingJobs(app)
		}
	}()

	log.Printf("✅ AI Queue Processor running (interval: %v)", interval)
}

// processPendingJobs processes all pending AI jobs
func processPendingJobs(app core.App) {
	// Find pending jobs ordered by priority (desc) and scheduled_at (asc)
	jobs, err := app.FindRecordsByFilter(
		"ai_processing_queue",
		"status = {:status}",
		"",
		"-priority,scheduled_at",
		10, // Process up to 10 jobs at a time
		0,
		nil,
		map[string]any{"status": "pending"},
	)

	if err != nil {
		log.Printf("Error finding pending jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Printf("📋 Found %d pending AI jobs", len(jobs))

	for _, job := range jobs {
		if err := processJob(app, job); err != nil {
			log.Printf("❌ Failed to process job %s: %v", job.Id, err)
		}
	}
}

// processJob processes a single AI job
func processJob(app core.App, job *core.Record) error {
	// Check if scheduled time has arrived
	scheduledAt, ok := job.DateTime("scheduled_at", time.UTC)
	if !ok {
		return markJobFailed(app, job, "Invalid scheduled_at time")
	}

	if time.Now().UTC().Before(scheduledAt) {
		// Not time yet, skip
		return nil
	}

	// Get estimated tokens
	estimatedTokens := float64(job.GetInt64Value("estimated_tokens"))
	if estimatedTokens == 0 {
		estimatedTokens = 1000 // Default estimate
	}

	// Check rate limit
	if !aiTokenBucket.Consume(estimatedTokens) {
		log.Printf("⏳ Rate limit reached, job %s will wait", job.Id)
		return nil
	}

	// Mark job as processing
	job.Set("status", "processing")
	job.Set("started_at", time.Now().UTC().Format(time.RFC3339))
	if err := app.Save(job); err != nil {
		return err
	}

	// Process based on job type
	jobType := job.GetStringValue("job_type")
	var err error

	switch jobType {
	case "entry_analysis":
		err = processEntryAnalysis(app, job)
	case "daily_summary":
		err = processDailySummary(app, job)
	case "weekly_analysis":
		err = processWeeklyAnalysis(app, job)
	case "monthly_analysis":
		err = processMonthlyAnalysis(app, job)
	case "streak_update":
		err = processStreakUpdate(app, job)
	case "growth_calculation":
		err = processGrowthCalculation(app, job)
	default:
		err = markJobFailed(app, job, "Unknown job type: "+jobType)
	}

	if err != nil {
		// Increment attempt count
		attempts := job.GetInt64Value("attempts")
		job.Set("attempts", attempts+1)

		// Max retries: 3
		if attempts >= 3 {
			return markJobFailed(app, job, err.Error())
		}

		// Re-queue with delay
		job.Set("status", "pending")
		job.Set("scheduled_at", time.Now().Add(5*time.Minute).UTC().Format(time.RFC3339))
		return app.Save(job)
	}

	// Mark job as completed
	return markJobCompleted(app, job)
}

// markJobCompleted marks a job as completed
func markJobCompleted(app core.App, job *core.Record) error {
	job.Set("status", "completed")
	job.Set("completed_at", time.Now().UTC().Format(time.RFC3339))
	if err := app.Save(job); err != nil {
		return err
	}
	log.Printf("✅ Job %s completed", job.Id)
	return nil
}

// markJobFailed marks a job as failed
func markJobFailed(app core.App, job *core.Record, errorMsg string) error {
	job.Set("status", "failed")
	job.Set("error_message", errorMsg)
	job.Set("completed_at", time.Now().UTC().Format(time.RFC3339))
	if err := app.Save(job); err != nil {
		return err
	}
	log.Printf("❌ Job %s failed: %s", job.Id, errorMsg)
	return nil
}

// Stub functions for actual AI processing (to be implemented in Phase 4)
// These are placeholder implementations for Phase 1

func processEntryAnalysis(app core.App, job *core.Record) error {
	log.Printf("🔍 Processing entry analysis for job %s", job.Id)

	// TODO: Phase 4 - Implement actual AI Studio integration
	// For now, just mark as completed

	// Steps for Phase 4:
	// 1. Fetch the journal entry
	// 2. Decrypt the content (requires client key - this is complex!)
	// 3. Send to AI Studio with entry analysis prompt
	// 4. Store result in growth_analysis collection
	// 5. Update entry's ai_processed flag

	return nil
}

func processDailySummary(app core.App, job *core.Record) error {
	log.Printf("📊 Processing daily summary for job %s", job.Id)
	// TODO: Phase 4
	return nil
}

func processWeeklyAnalysis(app core.App, job *core.Record) error {
	log.Printf("📈 Processing weekly analysis for job %s", job.Id)
	// TODO: Phase 4
	return nil
}

func processMonthlyAnalysis(app core.App, job *core.Record) error {
	log.Printf("📉 Processing monthly analysis for job %s", job.Id)
	// TODO: Phase 4
	return nil
}

func processStreakUpdate(app core.App, job *core.Record) error {
	log.Printf("🔥 Processing streak update for job %s", job.Id)
	// TODO: Phase 3 - Recalculate streaks
	return nil
}

func processGrowthCalculation(app core.App, job *core.Record) error {
	log.Printf("📈 Processing growth calculation for job %s", job.Id)
	// TODO: Phase 4
	return nil
}

// getEnvFloat gets an environment variable as float64
func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := parseFloat(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseFloat parses a string to float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
