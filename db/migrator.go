package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migrator struct {
	db      *sql.DB
	Migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, dbProvider string, dbName string) (*Migrator, error) {
	// Verify migrations exist
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no migration files found in embedded filesystem")
	}

	log.Printf("Found %d migration files", len(entries))

	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	var driver database.Driver

	switch dbProvider {
	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{})
	case "cockroachdb":
		driver, err = cockroachdb.WithInstance(db, &cockroachdb.Config{})
	case "sqlite3":
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	default:
		return nil, fmt.Errorf("failed to initialize database driver: %s it may not be supported yet", dbProvider)
	}

	if err != nil {
		log.Fatalf("Failed to create DB driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, dbName, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		db:      db,
		Migrate: m,
	}, nil
}

func (m *Migrator) MigrateIfNeeded() error {
	curVersion, dirty, err := m.Migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, manual intervention is required")
	}

	hasPending, nextVersion, err := m.hasPendingMigrations(curVersion)
	if err != nil {
		return fmt.Errorf("failed to check pending migrations: %w", err)
	}

	if !hasPending {
		log.Printf("Database is up to date (version: %d)", curVersion)
	}

	log.Printf("Migrating database from version %d to %d", curVersion, nextVersion)

	if err := m.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations %w", err)
	}

	newVersion, _, err := m.Migrate.Version()
	if err != nil {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	log.Printf("Database migrated successfully to version %d", newVersion)
	return nil
}

func (m *Migrator) hasPendingMigrations(curVersion uint) (bool, uint, error) {
	nextVersion, err := m.getNextVersion(curVersion)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return false, curVersion, nil
		}
		return false, 0, err
	}

	return nextVersion > curVersion, nextVersion, nil
}

func (m *Migrator) getNextVersion(curVersion uint) (uint, error) {
	err := m.Migrate.Steps(1)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, migrate.ErrShortLimit{}) {
			return curVersion, migrate.ErrNoChange
		}
		return 0, err
	}

	newVersion, _, err := m.Migrate.Version()
	if err != nil {
		return 0, err
	}

	if err := m.Migrate.Migrate(curVersion); err != nil {
		return 0, fmt.Errorf("failed to rollback to original version: %w", err)
	}

	return newVersion, nil
}

// Close closes the source and the database.
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.Migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}

	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}

	return nil
}
