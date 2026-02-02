package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// ================================================================
		// 1. Update Users Collection - Add journaling fields
		// ================================================================
		users, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// Add encryption key hash field (for client key verification)
		users.Fields.Add(&core.TextField{
			Name:   "encryption_key_hash",
			Hidden: true,
		})

		// Add journaling stats fields
		users.Fields.Add(&core.NumberField{
			Name: "total_entries",
		})
		users.Fields.Add(&core.NumberField{
			Name: "total_words",
		})
		users.Fields.Add(&core.NumberField{
			Name: "current_streak",
		})
		users.Fields.Add(&core.NumberField{
			Name: "longest_streak",
		})
		users.Fields.Add(&core.DateField{
			Name: "last_entry_date",
		})
		users.Fields.Add(&core.SelectField{
			Name:      "preferred_analysis_frequency",
			Values:    []string{"daily", "weekly", "monthly"},
			MaxSelect: 1,
		})

		if err := app.Save(users); err != nil {
			return err
		}

		// ================================================================
		// 2. Journal Entries Collection (Encrypted Content)
		// ================================================================
		entries := core.NewBaseCollection("journal_entries")

		// Owner-only access - users can only see their own entries
		entries.ListRule = types.Pointer("@request.auth.id = user.id")
		entries.ViewRule = types.Pointer("@request.auth.id = user.id")
		entries.CreateRule = types.Pointer("@request.auth.id = user.id")
		entries.UpdateRule = types.Pointer("@request.auth.id = user.id")
		entries.DeleteRule = types.Pointer("@request.auth.id = user.id")

		// User relation
		entries.Fields.Add(&core.RelationField{
			Name:          "user",
			CollectionId:  users.Id,
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
		})

		// Entry date (the date this entry represents)
		entries.Fields.Add(&core.DateField{
			Name:     "entry_date",
			Required: true,
		})

		// Encrypted content (AES-256 encrypted journal text)
		entries.Fields.Add(&core.TextField{
			Name:     "encrypted_content",
			Required: true,
		})

		// Content hash for integrity verification
		entries.Fields.Add(&core.TextField{
			Name: "content_hash",
		})

		// Mood rating (1-10, self-reported)
		entries.Fields.Add(&core.NumberField{
			Name: "mood_rating",
		})

		// Tags (JSON array of strings)
		entries.Fields.Add(&core.JSONField{
			Name: "tags",
		})

		// Word count (cleartext metadata for stats)
		entries.Fields.Add(&core.NumberField{
			Name: "word_count",
		})

		// Whether entry has been processed by AI
		entries.Fields.Add(&core.BoolField{
			Name: "ai_processed",
		})

		// Add indexes
		entries.AddIndex("idx_entries_user_date", false, "user,entry_date", "")
		entries.AddIndex("idx_entries_entry_date", false, "entry_date", "")

		if err := app.Save(entries); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: remove collections and user fields
		// Note: Cannot fully rollback user field removal in safe way
		// Just delete the collections we created

		if col, err := app.FindCollectionByNameOrId("journal_entries"); err == nil {
			app.Delete(col)
		}

		return nil
	})
}
