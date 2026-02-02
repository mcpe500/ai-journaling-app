package hooks

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RegisterEntryHooks registers all journal entry related hooks
func RegisterEntryHooks(app core.App) {
	// Hook: After journal entry is created
	app.OnRecordAfterCreateRequest("journal_entries").BindFunc(func(e *core.RecordHookEvent) error {
		record := e.Record

		// 1. Update user's journaling stats
		if err := updateUserStatsAfterEntry(app, record); err != nil {
			log.Printf("Warning: Failed to update user stats: %v", err)
		}

		// 2. Add AI processing job to queue
		if err := queueAIAnalysisJob(app, record); err != nil {
			log.Printf("Warning: Failed to queue AI job: %v", err)
		}

		// 3. Invalidate heatmap cache for affected month/year
		if err := invalidateHeatmapCache(app, record); err != nil {
			log.Printf("Warning: Failed to invalidate heatmap cache: %v", err)
		}

		return e.Next()
	})

	// Hook: After journal entry is updated
	app.OnRecordAfterUpdateRequest("journal_entries").BindFunc(func(e *core.RecordHookEvent) error {
		record := e.Record

		// Invalidate heatmap cache when entry is modified
		if err := invalidateHeatmapCache(app, record); err != nil {
			log.Printf("Warning: Failed to invalidate heatmap cache: %v", err)
		}

		// Re-queue AI analysis if content changed significantly
		// TODO: Add logic to detect if encrypted_content changed
		if err := queueAIAnalysisJob(app, record); err != nil {
			log.Printf("Warning: Failed to queue AI job: %v", err)
		}

		return e.Next()
	})

	// Hook: After journal entry is deleted
	app.OnRecordAfterDeleteRequest("journal_entries").BindFunc(func(e *core.RecordHookEvent) error {
		record := e.Record

		// 1. Update user stats (decrement counts)
		if err := updateUserStatsAfterDeletion(app, record); err != nil {
			log.Printf("Warning: Failed to update user stats after deletion: %v", err)
		}

		// 2. Invalidate heatmap cache
		if err := invalidateHeatmapCache(app, record); err != nil {
			log.Printf("Warning: Failed to invalidate heatmap cache: %v", err)
		}

		return e.Next()
	})
}

// updateUserStatsAfterEntry updates user statistics after a new entry is created
func updateUserStatsAfterEntry(app core.App, record *core.Record) error {
	userID := record.GetStringValue("user")
	if userID == "" {
		return nil
	}

	user, err := app.FindAuthUserById(userID)
	if err != nil {
		return err
	}

	// Get entry date
	entryDate, ok := record.DateTime("entry_date", time.UTC)
	if !ok {
		return nil
	}

	// Increment total entries
	totalEntries := user.GetInt64Value("total_entries")
	user.Set("total_entries", totalEntries+1)

	// Add word count
	wordCount := record.GetInt64Value("word_count")
	totalWords := user.GetInt64Value("total_words")
	user.Set("total_words", totalWords+wordCount)

	// Calculate and update streak
	if err := calculateAndUpdateStreak(app, user, entryDate); err != nil {
		log.Printf("Warning: Failed to calculate streak: %v", err)
	}

	// Update last entry date
	user.Set("last_entry_date", entryDate.Format("2006-01-02"))

	if err := app.Save(user); err != nil {
		return err
	}

	log.Printf("✅ Updated stats for user %s: entries=%d, words=%d", user.Email(), totalEntries+1, totalWords+wordCount)
	return nil
}

// updateUserStatsAfterDeletion updates user statistics after an entry is deleted
func updateUserStatsAfterDeletion(app core.App, record *core.Record) error {
	userID := record.GetStringValue("user")
	if userID == "" {
		return nil
	}

	user, err := app.FindAuthUserById(userID)
	if err != nil {
		return err
	}

	// Decrement total entries
	totalEntries := user.GetInt64Value("total_entries")
	user.Set("total_entries", max(0, totalEntries-1))

	// Subtract word count
	wordCount := record.GetInt64Value("word_count")
	totalWords := user.GetInt64Value("total_words")
	user.Set("total_words", max(0, totalWords-wordCount))

	// Recalculate streak from scratch (expensive but accurate)
	if err := recalculateStreak(app, user); err != nil {
		log.Printf("Warning: Failed to recalculate streak: %v", err)
	}

	if err := app.Save(user); err != nil {
		return err
	}

	log.Printf("✅ Updated stats for user %s after deletion", user.Email())
	return nil
}

// calculateAndUpdateStreak updates the writing streak based on the new entry
func calculateAndUpdateStreak(app core.App, user *core.Record, newEntryDate time.Time) error {
	lastEntryDateStr := user.GetStringValue("last_entry_date")
	var lastEntryDate time.Time
	var err error

	if lastEntryDateStr != "" {
		lastEntryDate, err = time.Parse("2006-01-02", lastEntryDateStr)
		if err != nil {
			// If parsing fails, treat as no previous entry
			lastEntryDate = time.Time{}
		}
	}

	currentStreak := user.GetInt64Value("current_streak")
	longestStreak := user.GetInt64Value("longest_streak")

	// Check if the new entry is consecutive day
	if !lastEntryDate.IsZero() {
		daysDiff := int(newEntryDate.Sub(lastEntryDate).Hours() / 24)

		if daysDiff == 1 {
			// Consecutive day - increment streak
			currentStreak++
		} else if daysDiff > 1 {
			// Streak broken - start new streak
			currentStreak = 1
		}
		// If daysDiff == 0, same day - don't change streak
		// If daysDiff < 0, backdated entry - handle specially
	} else {
		// First entry
		currentStreak = 1
	}

	// Update longest streak
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	user.Set("current_streak", currentStreak)
	user.Set("longest_streak", longestStreak)

	return nil
}

// recalculateStreak recalculates the entire streak from all user entries
func recalculateStreak(app core.App, user *core.Record) error {
	userID := user.Id()

	// Get all entries ordered by date
	entries, err := app.FindRecordsByFilter(
		"journal_entries",
		"user = {:userId}",
		"",
		"entry_date",
		50,
		0,
		nil,
		map[string]any{"userId": userID},
	)

	if err != nil || len(entries) == 0 {
		user.Set("current_streak", 0)
		user.Set("longest_streak", 0)
		return nil
	}

	currentStreak := 1
	longestStreak := 1
	lastDate, _ := entries[0].DateTime("entry_date", time.UTC)

	for i := 1; i < len(entries); i++ {
		entryDate, _ := entries[i].DateTime("entry_date", time.UTC)
		daysDiff := int(lastDate.Sub(entryDate).Hours() / 24)

		if daysDiff == 1 {
			currentStreak++
		} else if daysDiff > 1 {
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
			currentStreak = 1
		}

		lastDate = entryDate
	}

	// Final check
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	user.Set("current_streak", currentStreak)
	user.Set("longest_streak", longestStreak)

	return nil
}

// queueAIAnalysisJob adds an AI analysis job to the processing queue
func queueAIAnalysisJob(app core.App, record *core.Record) error {
	// Get collections
	queueCollection, err := app.FindCollectionByNameOrId("ai_processing_queue")
	if err != nil {
		return err
	}

	userID := record.GetStringValue("user")
	entryDate, _ := record.DateTime("entry_date", time.UTC)
	scheduledAt := time.Now().UTC()

	// Create the queue job
	job := core.NewRecord(queueCollection)
	job.Set("user", userID)
	job.Set("job_type", "entry_analysis")
	job.Set("entry_id", record.Id)
	job.Set("status", "pending")
	job.Set("priority", 5) // Medium priority
	job.Set("attempts", 0)
	job.Set("scheduled_at", scheduledAt)
	job.Set("estimated_tokens", 1000) // Estimate for single entry analysis

	if err := app.Save(job); err != nil {
		return err
	}

	log.Printf("✅ Queued AI analysis job for entry %s", record.Id)
	return nil
}

// invalidateHeatmapCache invalidates the heatmap cache for the affected period
func invalidateHeatmapCache(app core.App, record *core.Record) error {
	userID := record.GetStringValue("user")
	if userID == "" {
		return nil
	}

	entryDate, ok := record.DateTime("entry_date", time.UTC)
	if !ok {
		return nil
	}

	year := entryDate.Year()
	month := int(entryDate.Month())

	// Delete cache for this specific month
	cacheCollection, err := app.FindCollectionByNameOrId("calendar_heatmap_cache")
	if err != nil {
		return err
	}

	// Find and delete cache records
	caches, err := app.FindRecordsByFilter(
		"calendar_heatmap_cache",
		"user = {:userId} && year = {:year} && month = {:month}",
		"",
		"",
		10,
		0,
		nil,
		map[string]any{
			"userId": userID,
			"year":   year,
			"month":  month,
		},
	)

	for _, cache := range caches {
		if err := app.Delete(cache); err != nil {
			log.Printf("Warning: Failed to delete cache record: %v", err)
		}
	}

	log.Printf("✅ Invalidated heatmap cache for user %s, %d-%d", userID, year, month)
	return nil
}

// Helper function for max of int64
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
