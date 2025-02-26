package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, conString string) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, conString)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func (s *PostgresStorage) Ping() error {
	err := s.db.Ping(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) Close() {
	s.db.Close()
}
