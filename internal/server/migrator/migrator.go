package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Migrator struct {
	migrator *migrate.Migrate
	db       *sql.DB
}

func NewMigrator(pool *pgxpool.Pool, migrationsPath string) (*Migrator, error) {
	db := stdlib.OpenDBFromPool(pool)

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)

	if err != nil {
		return nil, err
	}

	return &Migrator{
		migrator: m,
		db:       db,
	}, nil
}

func (m *Migrator) InitializeDB() error {
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("failed to close db")
		}
	}(m.db)
	defer func(migrator *migrate.Migrate) {
		err, _ := migrator.Close()
		if err != nil {
			fmt.Println("failed to close migrator")
		}
	}(m.migrator)

	if err := m.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
