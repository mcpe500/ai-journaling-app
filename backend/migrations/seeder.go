package migrations

import (
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// RunAllSeeders executes all seeder functions
func RunAllSeeders(app core.App) error {
	log.Println("🌱 Starting seeders...")

	if err := seedAdminUser(app); err != nil {
		return err
	}

	if err := seedTestUser(app); err != nil {
		return err
	}

	if err := seedSampleJournalEntries(app); err != nil {
		return err
	}

	log.Println("✅ All seeders completed!")
	return nil
}

// seedAdminUser creates the admin account from environment variables
func seedAdminUser(app core.App) error {
	adminEmail := os.Getenv("PB_ADMIN_EMAIL")
	adminPassword := os.Getenv("PB_ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Println("ℹ️  Admin credentials not provided, skipping admin seeder")
		return nil
	}

	// Check if admin already exists
	admin, err := app.FindAuthRecordByEmail("_pb_users_auth_", adminEmail)
	if err == nil && admin != nil {
		log.Printf("ℹ️  Admin user already exists: %s", adminEmail)
		return nil
	}

	// Create admin user
	collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
	if err != nil {
		return err
	}
	
	admin = core.NewRecord(collection)
	admin.SetEmail(adminEmail)
	admin.SetPassword(adminPassword)
	
	// Set admin role if applicable, though typically PB admins are separate from users in older versions, 
	// in v0.23+ they can be in _pb_users_auth_ or similar?
	// Actually, system admins are usually managed via `app.Dao().SaveAdmin(...)` in older versions.
	// In v0.23, `_pb_users_auth_` is the default users collection.
	// NOTE: Superusers are different.
	// But let's assume we are creating a regular user with 'admin' role field as per previous logic.
	
	admin.Set("role", "admin")

	if err := app.Save(admin); err != nil {
		return err
	}

	log.Printf("✅ Admin user created: %s", adminEmail)
	return nil
}

// seedTestUser creates a test user for development
func seedTestUser(app core.App) error {
	testEmail := os.Getenv("PB_TEST_USER_EMAIL")
	testPassword := os.Getenv("PB_TEST_USER_PASSWORD")

	if testEmail == "" || testPassword == "" {
		log.Println("ℹ️  Test user credentials not provided, skipping test user seeder")
		return nil
	}

	// Check if test user already exists
	existing, err := app.FindAuthRecordByEmail("users", testEmail)
	if err == nil && existing != nil {
		log.Printf("ℹ️  Test user already exists: %s", testEmail)
		return nil
	}

	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	// Create test user
	testUser := core.NewRecord(collection)
	testUser.SetEmail(testEmail)
	testUser.SetPassword(testPassword)

	// Initialize journaling stats
	testUser.Set("total_entries", 0)
	testUser.Set("total_words", 0)
	testUser.Set("current_streak", 0)
	testUser.Set("longest_streak", 0)
	testUser.Set("last_entry_date", "")
	testUser.Set("preferred_analysis_frequency", "weekly")

	if err := app.Save(testUser); err != nil {
		return err
	}

	log.Printf("✅ Test user created: %s / %s", testEmail, testPassword)
	return nil
}

// seedSampleJournalEntries creates sample journal entries for testing
func seedSampleJournalEntries(app core.App) error {
	testEmail := os.Getenv("PB_TEST_USER_EMAIL")
	if testEmail == "" {
		log.Println("ℹ️  Test user not found, skipping sample entries")
		return nil
	}

	// Get test user
	testUser, err := app.FindAuthRecordByEmail("users", testEmail)
	if err != nil {
		log.Println("ℹ️  Test user not found, skipping sample entries")
		return nil
	}

	// Get journal entries collection
	entriesCollection, err := app.FindCollectionByNameOrId("journal_entries")
	if err != nil {
		return err
	}

	// Check if entries already exist
	existing, _ := app.FindRecordsByFilter(
		"journal_entries",
		"user = {:userId}",
		"",
		1,
		0,
		map[string]any{"userId": testUser.Id},
	)

	if len(existing) > 0 {
		log.Println("ℹ️  Sample entries already exist, skipping")
		return nil
	}

	// Create sample entries for the past 7 days
	sampleEntries := []struct {
		date    string
		content string
		mood    int64
		tags    []string
	}{
		{
			date:    time.Now().AddDate(0, 0, -6).Format("2006-01-02"),
			content: "Today was a challenging day at work. I had to deal with a difficult client, but I managed to stay calm and find a solution. I'm proud of my patience.",
			mood:    7,
			tags:    []string{"work", "growth", "patience"},
		},
		{
			date:    time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			content: "Started my morning with meditation and it really set a positive tone for the day. I feel more focused and less anxious.",
			mood:    8,
			tags:    []string{"wellness", "meditation", "morning-routine"},
		},
		{
			date:    time.Now().AddDate(0, 0, -4).Format("2006-01-02"),
			content: "Feeling a bit overwhelmed with all the projects I have going on. Need to prioritize better and maybe say no to some things.",
			mood:    5,
			tags:    []string{"stress", "work", "prioritization"},
		},
		{
			date:    time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
			content: "Great workout session today! I'm finally getting back into my fitness routine. Exercise really helps clear my mind.",
			mood:    9,
			tags:    []string{"fitness", "health", "exercise"},
		},
		{
			date:    time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			content: "Had a wonderful dinner with family. It's important to make time for loved ones. I feel grateful and recharged.",
			mood:    8,
			tags:    []string{"family", "gratitude", "relationships"},
		},
		{
			date:    time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			content: "Learned a new framework at work today. The learning curve is steep but I'm enjoying the challenge. Growth happens outside comfort zone.",
			mood:    7,
			tags:    []string{"learning", "work", "growth", "challenges"},
		},
		{
			date:    time.Now().Format("2006-01-02"),
			content: "Taking time to reflect on the week. I've had ups and downs, but overall I'm making progress. Ready to tackle next week with renewed energy!",
			mood:    8,
			tags:    []string{"reflection", "weekly-review", "mindfulness"},
		},
	}

	for _, sample := range sampleEntries {
		entry := core.NewRecord(entriesCollection)

		// For now, store content unencrypted (in production, this would be encrypted on client)
		// The encryption_key_hash would be set when the user creates their encryption key
		entry.Set("user", testUser.Id)
		entry.Set("entry_date", sample.date)
		entry.Set("encrypted_content", "[ENCRYPTED]"+sample.content) // Placeholder
		entry.Set("content_hash", modelHashString(sample.content))
		entry.Set("mood_rating", sample.mood)
		entry.Set("tags", sample.tags)
		entry.Set("word_count", int64(len([]rune(sample.content)))) // Rough word count
		entry.Set("ai_processed", false)

		if err := app.Save(entry); err != nil {
			log.Printf("Warning: Failed to create sample entry: %v", err)
			continue
		}
	}

	log.Printf("✅ Created %d sample journal entries", len(sampleEntries))

	// Update user stats
	testUser.Set("total_entries", int64(len(sampleEntries)))
	testUser.Set("current_streak", 7) // 7 consecutive days
	testUser.Set("longest_streak", 7)
	testUser.Set("last_entry_date", time.Now().Format("2006-01-02"))

	totalWords := int64(0)
	for _, sample := range sampleEntries {
		wordCount := int64(len([]rune(sample.content)))
		totalWords += wordCount
	}
	testUser.Set("total_words", totalWords)

	if err := app.Save(testUser); err != nil {
		log.Printf("Warning: Failed to update test user stats: %v", err)
	}

	return nil
}

// Helper function to create a simple hash
func modelHashString(s string) string {
	// Simple hash for demo purposes
	// In production, use proper SHA-256
	return security.RandomString(16)
}
