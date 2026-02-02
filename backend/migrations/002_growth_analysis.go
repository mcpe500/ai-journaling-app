package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Get the users collection for relation
		users, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// ================================================================
		// Growth Analysis Collection (AI-Generated Insights)
		// ================================================================
		growthAnalysis := core.NewBaseCollection("growth_analysis")

		// Owner-only access
		growthAnalysis.ListRule = types.Pointer("@request.auth.id = user.id")
		growthAnalysis.ViewRule = types.Pointer("@request.auth.id = user.id")
		growthAnalysis.CreateRule = nil // Backend only (AI creates these)
		growthAnalysis.UpdateRule = nil // Backend only
		growthAnalysis.DeleteRule = types.Pointer("@request.auth.id = user.id")

		// User relation
		growthAnalysis.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		})

		// Analysis type (daily/weekly/monthly)
		growthAnalysis.Fields.Add(&core.SelectField{
			Name:     "analysis_type",
			Values:   []string{"entry", "daily", "weekly", "monthly"},
			Required: true,
			MaxSelect: 1,
		})

		// Period start date
		growthAnalysis.Fields.Add(&core.DateField{
			Name:     "period_start",
			Required: true,
		})

		// Period end date
		growthAnalysis.Fields.Add(&core.DateField{
			Name:     "period_end",
			Required: true,
		})

		// Encrypted insights (AI-generated content, encrypted for privacy)
		growthAnalysis.Fields.Add(&core.TextField{
			Name:     "encrypted_insights",
			Required: false,
		})

		// Growth score (0-100, calculated by AI)
		growthAnalysis.Fields.Add(&core.NumberField{
			Name: "growth_score",
		})

		// Key themes (JSON array of recurring topics)
		growthAnalysis.Fields.Add(&core.JSONField{
			Name: "key_themes",
		})

		// Emotional trend (improving/stable/declining)
		growthAnalysis.Fields.Add(&core.SelectField{
			Name:   "emotional_trend",
			Values: []string{"improving", "stable", "declining"},
			MaxSelect: 1,
		})

		// Action items (JSON array of AI-suggested actions)
		growthAnalysis.Fields.Add(&core.JSONField{
			Name: "action_items",
		})

		// Motivational quote (AI-selected)
		growthAnalysis.Fields.Add(&core.TextField{
			Name: "motivation_quote",
		})

		// Related journal entries (many-to-many relation)
		growthAnalysis.Fields.Add(&core.RelationField{
			Name:     "related_entries",
			MaxSelect: 100, // Support weekly/monthly analyses
		})

		// Add indexes for efficient queries
		growthAnalysis.AddIndex("idx_analysis_user_period", false, "user,analysis_type,period_start", "")

		if err := app.Save(growthAnalysis); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: delete the collection
		if col, err := app.FindCollectionByNameOrId("growth_analysis"); err == nil {
			app.Delete(col)
		}
		return nil
	})
}
