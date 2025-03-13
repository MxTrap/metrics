package postgres_storage

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresStorage struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

type dbMetric struct {
	ID    int64    `db:"id"`
	MType string   `db:"metric_type"`
	Name  string   `db:"metric_name"`
	Value *float64 `db:"value"`
	Delta *int64   `db:"delta"`
}

func NewPostgresStorage(db *pgxpool.Pool, log *logger.Logger) (*PostgresStorage, error) {

	return &PostgresStorage{
		db:  db,
		log: log,
	}, nil
}

func (*PostgresStorage) mapCommonToDBMetric(metric models.Metric) dbMetric {
	return dbMetric{
		MType: metric.MType,
		Name:  metric.ID,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (*PostgresStorage) mapDBToCommonMetric(metric dbMetric) models.Metric {
	return models.Metric{
		ID:    metric.Name,
		MType: metric.MType,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (*PostgresStorage) withRetry(cb func() error) error {
	const maxRetryAmount = 3
	for i := 0; i <= maxRetryAmount; i++ {
		err := cb()
		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
			return err
		}

		if i < maxRetryAmount {
			time.Sleep(time.Duration(1+2*i) * time.Second)
		}
	}

	return nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	if s.db == nil {
		return errors.New("database not initialized")
	}
	s.log.Logger.Info("Ping")
	return s.db.Ping(ctx)
}

func (s *PostgresStorage) Save(ctx context.Context, metric models.Metric) error {
	s.log.Logger.Info("Save")

	return s.withRetry(func() error {
		tx, err := s.db.Begin(ctx)

		if err != nil {
			return err
		}

		exec, err := tx.Exec(ctx, updateStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
		if err != nil {
			return err
		}
		if exec.RowsAffected() == 0 {
			_, err := tx.Exec(ctx, insertStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
			if err != nil {
				err := tx.Rollback(ctx)
				if err != nil {
					return err
				}
				return err
			}
		}
		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *PostgresStorage) Find(ctx context.Context, metricName string) (models.Metric, error) {
	s.log.Logger.Info("Find")

	var metric models.Metric

	err := s.withRetry(func() error {
		rows, err := s.db.Query(
			ctx,
			findStmt, metricName,
		)

		if err != nil {
			return err
		}

		defer rows.Close()

		m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbMetric])
		if err != nil {
			return err
		}

		metric = s.mapDBToCommonMetric(m)
		return nil
	})

	if err != nil {
		return models.Metric{}, err
	}

	return metric, nil
}

func (s *PostgresStorage) GetAll(ctx context.Context) (map[string]models.Metric, error) {
	s.log.Logger.Info("Get all")

	var metrics map[string]models.Metric

	err := s.withRetry(func() error {
		rows, err := s.db.Query(ctx, selectAllStmt)
		if err != nil {
			return err
		}
		defer rows.Close()
		dbMetrics, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbMetric])
		if err != nil {
			return err
		}
		cMetrics := make(map[string]models.Metric, len(dbMetrics))
		for _, m := range dbMetrics {
			cMetrics[m.Name] = s.mapDBToCommonMetric(m)
		}
		metrics = cMetrics
		return nil
	})
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *PostgresStorage) SaveAll(ctx context.Context, metrics map[string]models.Metric) error {
	s.log.Logger.Info("Save all")

	return s.withRetry(func() error {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}

		batchUpdate := pgx.Batch{}
		for _, metric := range metrics {
			batchUpdate.Queue(updateStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
		}

		batchResult := tx.SendBatch(ctx, &batchUpdate)

		insertRows := make([]models.Metric, 0, len(metrics))

		for _, metric := range metrics {
			row, err := batchResult.Exec()
			if err != nil {
				insertRows = append(insertRows, metric)
			}
			if row.RowsAffected() == 0 {
				insertRows = append(insertRows, metric)
			}
		}

		err = batchResult.Close()
		if err != nil {
			return err
		}

		if len(insertRows) > 0 {
			insertBatch := pgx.Batch{}

			for _, metric := range insertRows {
				insertBatch.Queue(insertStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
			}

			batchResult = tx.SendBatch(ctx, &insertBatch)

			for range insertRows {
				_, err := batchResult.Exec()
				if err != nil {
					err = tx.Rollback(ctx)
					if err != nil {
						return err
					}
					return err
				}
			}
			err := batchResult.Close()
			if err != nil {
				return err
			}

		}

		err = tx.Commit(ctx)

		if err != nil {
			return err
		}

		return nil
	})
}

func (s *PostgresStorage) Close() {
	s.db.Close()
}
