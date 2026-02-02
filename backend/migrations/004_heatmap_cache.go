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
		// Calendar Heatmap Cache Collection (Performance Optimization)
		// ================================================================
		heatmapCache := core.NewBaseCollection("calendar_heatmap_cache")

		// Owner-only access
		heatmapCache.ListRule = types.Pointer("@request.auth.id = user.id")
		heatmapCache.ViewRule = types.Pointer("@request.auth.id = user.id")
		heatmapCache.CreateRule = nil // Backend only (auto-generated)
		heatmapCache.UpdateRule = nil // Backend only
		heatmapCache.DeleteRule = nil // Backend only (auto-invalidated)

		// User relation
		heatmapCache.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		})

		// Year (e.g., 2026)
		heatmapCache.Fields.Add(&core.NumberField{
			Name:     "year",
			Required: true,
		})

		// Month (1-12, or 0 for full year view)
		heatmapCache.Fields.Add(&core.NumberField{
			Name:     "month",
			Required: true,
		})

		// Pre-computed heatmap data (JSON)
		// Structure: {"days": [{"date": "2026-01-01", "count": 2, "mood": 7, "color": "#4caf50"}, ...]}
		heatmapCache.Fields.Add(&core.JSONField{
			Name:     "data_json",
			Required: true,
		})

		// Last entry ID included in cache (for cache invalidation)
		heatmapCache.Fields.Add(&core.TextField{
			Name: "last_entry_id",
		})

		// Cache version (increment when entry added/modified)
		heatmapCache.Fields.Add(&core.NumberField{
			Name: "cache_version",
		})

		// Total entries in this period
		heatmapCache.Fields.Add(&core.NumberField{
			Name: "total_entries",
		})

		// Average mood for this period
		heatmapCache.Fields.Add(&core.NumberField{
			Name: "average_mood",
		})

		// Add unique index on user + year + month
		heatmapCache.AddIndex("idx_heatmap_user_year_month", true, "user,year,month", "")

		if err := app.Save(heatmapCache); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: delete the collection
		if col, err := app.FindCollectionByNameOrId("calendar_heatmap_cache"); err == nil {
			app.Delete(col)
		}
		return nil
	})
}
