package commands

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	"flora-hive/lib"
)

// MigrateUpCommand runs migrations up.
type MigrateUpCommand struct{}

func (s *MigrateUpCommand) Short() string            { return "migrate up the database" }
func (s *MigrateUpCommand) Setup(cmd *cobra.Command) {}
func (s *MigrateUpCommand) Run() lib.CommandRunner {
	return func(logger lib.Logger, database lib.Database) {
		logger.Info("Starting migration up run")
		m := createMigrationStruct(database, logger)
		err := m.Up()
		if err != nil && err.Error() != "no change" {
			logger.Fatal(fmt.Errorf("error when trying to migrate up: %w", err))
		}
		logger.Info("Migration up was successful")
	}
}

// NewMigrateUpCommand constructs MigrateUpCommand.
func NewMigrateUpCommand() *MigrateUpCommand { return &MigrateUpCommand{} }

// MigrateDownCommand runs migrations down one step.
type MigrateDownCommand struct{}

func (s *MigrateDownCommand) Short() string            { return "migrate down the database" }
func (s *MigrateDownCommand) Setup(cmd *cobra.Command) {}
func (s *MigrateDownCommand) Run() lib.CommandRunner {
	return func(logger lib.Logger, database lib.Database) {
		logger.Info("Starting migration down run")
		m := createMigrationStruct(database, logger)
		err := m.Down()
		if err != nil && err.Error() != "no change" {
			logger.Fatal(fmt.Errorf("error when trying to migrate down: %w", err))
		}
		logger.Info("Migration down was successful")
	}
}

// NewMigrateDownCommand constructs MigrateDownCommand.
func NewMigrateDownCommand() *MigrateDownCommand { return &MigrateDownCommand{} }

func createMigrationStruct(database lib.Database, logger lib.Logger) *migrate.Migrate {
	driver, err := postgres.WithInstance(database.DB.DB, &postgres.Config{})
	if err != nil {
		logger.Fatal(fmt.Errorf("error when trying to open database driver: %w", err))
	}
	migrationStruct, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		logger.Fatal(fmt.Errorf("error when trying to create migration struct: %w", err))
	}
	return migrationStruct
}
