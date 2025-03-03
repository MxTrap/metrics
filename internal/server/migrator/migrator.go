package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	migrator *migrate.Migrate
}

func NewMigrator(connString string) (*Migrator, error) {
	fmt.Println("conn string ", connString)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", utils.GetProjectPath()+"/migrations"),
		"postgres", driver)

	if err != nil {
		return nil, err
	}

	return &Migrator{
		migrator: m,
	}, nil
}

func (m *Migrator) InitializeDB() error {
	if err := m.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
