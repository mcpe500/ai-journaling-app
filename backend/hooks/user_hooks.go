package hooks

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterUserHooks registers all user-related hooks
func RegisterUserHooks(app core.App) {
	// Hook: After user is created (registration)
	app.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		user := e.Record

		// Initialize journaling stats to 0
		user.Set("total_entries", 0)
		user.Set("total_words", 0)
		user.Set("current_streak", 0)
		user.Set("longest_streak", 0)
		user.Set("last_entry_date", "")

		// Set default analysis frequency to weekly
		user.Set("preferred_analysis_frequency", "weekly")

		// Save the updated user
		if err := app.Save(user); err != nil {
			log.Printf("Warning: Failed to initialize user stats: %v", err)
			return err
		}

		log.Printf("✅ Initialized journaling stats for new user %s", user.Email())
		return e.Next()
	})

	// Hook: Before user is updated
	app.OnRecordUpdate("users").BindFunc(func(e *core.RecordEvent) error {
		// Prevent direct manipulation of stats (should only be updated via hooks)
		// This is a basic security measure - more validation can be added
		_ = e.Record // Suppress unused variable warning - placeholder for future validation

		return e.Next()
	})
}

// GetUserStats retrieves formatted statistics for a user
func GetUserStats(app core.App, userID string) (map[string]interface{}, error) {
	user, err := app.FindRecordById("users", userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_entries":      user.GetInt("total_entries"),
		"total_words":        user.GetInt("total_words"),
		"current_streak":     user.GetInt("current_streak"),
		"longest_streak":     user.GetInt("longest_streak"),
		"last_entry_date":    user.GetString("last_entry_date"),
		"analysis_frequency": user.GetString("preferred_analysis_frequency"),
	}

	return stats, nil
}
