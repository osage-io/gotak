package database

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version     int64
	Name        string
	UpScript    string
	DownScript  string
	AppliedAt   time.Time
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, logger *logger.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// RunMigrations applies all pending migrations
func (m *MigrationManager) RunMigrations() error {
	// Ensure migrations table exists
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migrations from embedded files
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get current migration version
	currentVersion, err := m.getCurrentMigrationVersion()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	m.logger.Info().
		Int64("current_version", currentVersion).
		Int("available_migrations", len(migrations)).
		Msg("Starting database migrations")

	// Apply pending migrations
	applied := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		if err := m.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %d (%s): %w", 
				migration.Version, migration.Name, err)
		}

		applied++
		m.logger.Info().
			Int64("version", migration.Version).
			Str("name", migration.Name).
			Msg("Applied migration")
	}

	if applied == 0 {
		m.logger.Info().Msg("Database is up to date, no migrations applied")
	} else {
		m.logger.Info().
			Int("applied_migrations", applied).
			Msg("Database migrations completed successfully")
	}

	return nil
}

// RollbackMigration rolls back the last migration
func (m *MigrationManager) RollbackMigration() error {
	currentVersion, err := m.getCurrentMigrationVersion()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Load migrations to find the one to rollback
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.Version == currentVersion {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration version %d not found", currentVersion)
	}

	if targetMigration.DownScript == "" {
		return fmt.Errorf("migration %d (%s) has no down script", 
			targetMigration.Version, targetMigration.Name)
	}

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute down script
	if _, err := tx.Exec(targetMigration.DownScript); err != nil {
		return fmt.Errorf("failed to execute down script: %w", err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", 
		targetMigration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	m.logger.Info().
		Int64("version", targetMigration.Version).
		Str("name", targetMigration.Name).
		Msg("Rolled back migration")

	return nil
}

// GetMigrationStatus returns the current migration status
func (m *MigrationManager) GetMigrationStatus() ([]Migration, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := m.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Mark applied migrations
	for i := range migrations {
		if appliedAt, exists := appliedVersions[migrations[i].Version]; exists {
			migrations[i].AppliedAt = appliedAt
		}
	}

	return migrations, nil
}

// createMigrationsTable creates the migrations tracking table
func (m *MigrationManager) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// loadMigrations loads migrations from embedded files
func (m *MigrationManager) loadMigrations() ([]Migration, error) {
	files, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		migration, err := m.parseMigrationFile(file.Name())
		if err != nil {
			m.logger.Warn().
				Str("file", file.Name()).
				Err(err).
				Msg("Failed to parse migration file, skipping")
			continue
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file
func (m *MigrationManager) parseMigrationFile(filename string) (Migration, error) {
	// Parse version and name from filename (e.g., "001_create_users.up.sql")
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.SplitN(name, "_", 2)
	
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}

	// Remove .up or .down suffix
	migrationName := parts[1]
	if strings.HasSuffix(migrationName, ".up") {
		migrationName = strings.TrimSuffix(migrationName, ".up")
	} else if strings.HasSuffix(migrationName, ".down") {
		migrationName = strings.TrimSuffix(migrationName, ".down")
	}

	// Read file content
	content, err := migrationFS.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read migration file %s: %w", filename, err)
	}

	migration := Migration{
		Version: version,
		Name:    migrationName,
	}

	// Determine if this is an up or down migration
	if strings.Contains(filename, ".up.") {
		migration.UpScript = string(content)
	} else if strings.Contains(filename, ".down.") {
		migration.DownScript = string(content)
	} else {
		// Default to up migration
		migration.UpScript = string(content)
	}

	return migration, nil
}

// getCurrentMigrationVersion gets the latest applied migration version
func (m *MigrationManager) getCurrentMigrationVersion() (int64, error) {
	var version sql.NullInt64
	err := m.db.QueryRow(`
		SELECT MAX(version) 
		FROM schema_migrations 
		WHERE NOT dirty
	`).Scan(&version)
	
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get current migration version: %w", err)
	}

	if !version.Valid {
		return 0, nil
	}

	return version.Int64, nil
}

// getAppliedMigrations returns a map of applied migration versions and their applied timestamps
func (m *MigrationManager) getAppliedMigrations() (map[int64]time.Time, error) {
	rows, err := m.db.Query(`
		SELECT version, applied_at 
		FROM schema_migrations 
		WHERE NOT dirty 
		ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]time.Time)
	for rows.Next() {
		var version int64
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		applied[version] = appliedAt
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return applied, nil
}

// runMigration applies a single migration
func (m *MigrationManager) runMigration(migration Migration) error {
	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Mark migration as dirty (in progress)
	if _, err := tx.Exec(`
		INSERT INTO schema_migrations (version, dirty) 
		VALUES ($1, true)
		ON CONFLICT (version) DO UPDATE SET dirty = true
	`, migration.Version); err != nil {
		return fmt.Errorf("failed to mark migration as dirty: %w", err)
	}

	// Execute migration script
	if _, err := tx.Exec(migration.UpScript); err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	// Mark migration as clean (completed)
	if _, err := tx.Exec(`
		UPDATE schema_migrations 
		SET dirty = false, applied_at = NOW() 
		WHERE version = $1
	`, migration.Version); err != nil {
		return fmt.Errorf("failed to mark migration as clean: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}
