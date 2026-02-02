package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get collections for relations
		users, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// journal_entries may not exist in rollback scenarios, so handle gracefully
		var entries *core.Collection
		if col, err := app.FindCollectionByNameOrId("journal_entries"); err == nil {
			entries = col
		}

		// ================================================================
		// AI Processing Queue Collection (Async Job Management)
		// ================================================================
		aiQueue := core.NewBaseCollection("ai_processing_queue")

		// Admin-only access (backend manages this)
		aiQueue.ListRule = nil
		aiQueue.ViewRule = nil
		aiQueue.CreateRule = nil
		aiQueue.UpdateRule = nil
		aiQueue.DeleteRule = nil

		// User relation (who triggered this job)
		aiQueue.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		})

		// Job type
		aiQueue.Fields.Add(&core.SelectField{
			Name:     "job_type",
			Values:   []string{"entry_analysis", "daily_summary", "weekly_analysis", "monthly_analysis", "streak_update", "growth_calculation"},
			Required: true,
			MaxSelect: 1,
		})

		// Related entry (nullable - not all jobs have an entry)
		if entries != nil {
			aiQueue.Fields.Add(&core.RelationField{
				Name:     "entry_id",
				CollectionId: entries.Id,
				Required: false,
				MaxSelect:    1,
			})
		}

		// Job status
		aiQueue.Fields.Add(&core.SelectField{
			Name:     "status",
			Values:   []string{"pending", "processing", "completed", "failed"},
			Required: true,
			MaxSelect: 1,
		})

		// Priority (1-10, higher = more urgent)
		aiQueue.Fields.Add(&core.NumberField{
			Name:     "priority",
			Required: true,
		})

		// Retry count
		aiQueue.Fields.Add(&core.NumberField{
			Name:     "attempts",
			Required: true,
		})

		// Error message (on failure)
		aiQueue.Fields.Add(&core.TextField{
			Name: "error_message",
		})

		// When to process this job
		aiQueue.Fields.Add(&core.DateField{
			Name:     "scheduled_at",
			Required: true,
		})

		// Processing started timestamp
		aiQueue.Fields.Add(&core.DateField{
			Name: "started_at",
		})

		// Processing completed timestamp
		aiQueue.Fields.Add(&core.DateField{
			Name: "completed_at",
		})

		// Estimated tokens needed (for rate limiting)
		aiQueue.Fields.Add(&core.NumberField{
			Name: "estimated_tokens",
		})

		// Add indexes for efficient job polling
		aiQueue.AddIndex("idx_queue_status_scheduled", false, "status,scheduled_at", "")
		aiQueue.AddIndex("idx_queue_priority", false, "priority", "")

		if err := app.Save(aiQueue); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: delete the collection
		if col, err := app.FindCollectionByNameOrId("ai_processing_queue"); err == nil {
			app.Delete(col)
		}
		return nil
	})
}
