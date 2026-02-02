package hooks

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterUserHooks registers all user-related hooks
func RegisterUserHooks(app core.App) {
	// Hook: After user is created (registration)
	app.OnRecordAfterCreateRequest("_pb_users_auth_").BindFunc(func(e *core.RecordHookEvent) error {
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
	app.OnRecordBeforeUpdateRequest("_pb_users_auth_").BindFunc(func(e *core.RecordHookEvent) error {
		// Prevent direct manipulation of stats (should only be updated via hooks)
		// This is a basic security measure - more validation can be added
		user := e.Record
		original := e.Original

		// Check if someone is trying to manually modify stats
		if user.GetInt64Value("total_entries") != original.GetInt64Value("total_entries") &&
			e.HttpContext.Request().Header.Get("User-Agent") != "" {
			// Allow if it's from internal system (check via custom header or similar)
			// For now, we'll just log it
			log.Printf("⚠️  Direct modification of total_entries detected for user %s", user.Email())
		}

		return e.Next()
	})
}

// GetUserStats retrieves formatted statistics for a user
func GetUserStats(app core.App, userID string) (map[string]interface{}, error) {
	user, err := app.FindAuthUserById(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_entries":       user.GetInt64Value("total_entries"),
		"total_words":         user.GetInt64Value("total_words"),
		"current_streak":      user.GetInt64Value("current_streak"),
		"longest_streak":      user.GetInt64Value("longest_streak"),
		"last_entry_date":     user.GetStringValue("last_entry_date"),
		"analysis_frequency":  user.GetStringValue("preferred_analysis_frequency"),
	}

	return stats, nil
}
