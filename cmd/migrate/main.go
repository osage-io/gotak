package main

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/dfedick/gotak/internal/database"
	"github.com/dfedick/gotak/pkg/logger"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	var (
		dbURL = flag.String("db", "postgres://gotak:gotak_dev_password@localhost:5432/gotak_dev?sslmode=disable", "database URL")
		cmd   = flag.String("cmd", "up", "migration command: up, down, status")
	)
	flag.Parse()

	// Initialize logger
	loggerConfig := logger.Config{
		Level:      "info",
		Format:     "console",
		Output:     "stdout",
		Service:    "migration",
		Version:    "1.0.0",
		TimeFormat: "rfc3339",
	}
	logger.Initialize(loggerConfig)
	log := logger.GetGlobalLogger()

	// Connect to database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Database connection failed")
	}

	log.Info().Str("database", *dbURL).Msg("Connected to database")

	// Create migration manager
	migrator := database.NewMigrationManager(db, log)

	// Execute command
	switch *cmd {
	case "up":
		if err := migrator.RunMigrations(); err != nil {
			log.Fatal().Err(err).Msg("Migration failed")
		}
		log.Info().Msg("Migrations completed successfully")

	case "down":
		if err := migrator.RollbackMigration(); err != nil {
			log.Fatal().Err(err).Msg("Rollback failed")
		}
		log.Info().Msg("Rollback completed successfully")

	case "status":
		migrations, err := migrator.GetMigrationStatus()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get migration status")
		}

		fmt.Println("Migration Status:")
		fmt.Println("=================")
		for _, migration := range migrations {
			status := "Pending"
			appliedAt := ""
			if !migration.AppliedAt.IsZero() {
				status = "Applied"
				appliedAt = migration.AppliedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("Version: %d | Name: %s | Status: %s | Applied: %s\n",
				migration.Version, migration.Name, status, appliedAt)
		}

	default:
		log.Fatal().Str("command", *cmd).Msg("Unknown command. Use: up, down, or status")
	}
}
