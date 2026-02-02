package main

import (
	"log"
	"os"
	"strings"

	"ai-journal-backend/hooks"
	"ai-journal-backend/migrations"
	_ "ai-journal-backend/migrations"
	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	app := pocketbase.New()

	// Check if running via go run (development mode)
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	// Allow automigrate via env var
	autoMigrate := isGoRun || os.Getenv("PB_AUTOMIGRATE") == "true"

	// Configure server bind address
	if httpAddr := os.Getenv("PB_HTTP"); httpAddr != "" {
		app.RootCmd.PersistentFlags().String("http", httpAddr, "Server address")
	}

	// Allow running seeders via environment variable
	runSeeders := os.Getenv("PB_RUN_SEEDERS") == "true"

	// Enable AI queue processor via environment variable
	runAIQueue := os.Getenv("ENABLE_AI_QUEUE") == "true"

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: autoMigrate,
	})

	// Register hooks for collections
	hooks.RegisterEntryHooks(app)
	hooks.RegisterUserHooks(app)
	log.Println("✅ Hooks registered successfully!")

	// Run seeders and start background services after app starts
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Run seeders if enabled
		if runSeeders {
			if err := migrations.RunAllSeeders(app); err != nil {
				log.Printf("Warning: Failed to run seeders: %v", err)
			} else {
				log.Println("✅ Seeders completed successfully!")
			}
		} else {
			log.Println("ℹ️  Seeders skipped. Set PB_RUN_SEEDERS=true to enable.")
		}

		// Start AI queue processor if enabled
		if runAIQueue {
			migrations.StartAIQueueProcessor(app)
			log.Println("✅ AI Queue Processor started!")
		} else {
			log.Println("ℹ️  AI Queue Processor skipped. Set ENABLE_AI_QUEUE=true to enable.")
		}

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
