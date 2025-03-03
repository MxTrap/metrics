package migrator

import (
	_ "database/sql"
	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"fmt"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	migrator *migrate.Migrate
}

func NewMigrator(connString string) (*Migrator, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", utils.GetProjectPath()+"/migrations"),
		connString,
	)

	if err != nil {
		return nil, err
	}

	return &Migrator{
		migrator: m,
	}, nil
}

func (m *Migrator) InitializeDB() error {
	if err := m.migrator.Up(); err != nil {
		return err
	}
	return nil
}
