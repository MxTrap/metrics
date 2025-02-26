package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, conString string) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, conString)
	if err != nil {
		return &PostgresStorage{}, err
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func (s *PostgresStorage) Ping() error {
	if s.db == nil {
		return errors.New("database not initialized")
	}
	return s.db.Ping(context.Background())
}

func (s *PostgresStorage) Close() {
	s.db.Close()
}
