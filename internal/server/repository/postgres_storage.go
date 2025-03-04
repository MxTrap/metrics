package repository

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
	db      *pgxpool.Pool
	connStr string
	log     *logger.Logger
}

type dbMetric struct {
	ID    int64    `db:"id"`
	MType string   `db:"metric_type"`
	Name  string   `db:"metric_name"`
	Value *float64 `db:"value"`
	Delta *int64   `db:"delta"`
}

func NewPostgresStorage(ctx context.Context, conString string, log *logger.Logger) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, conString)
	if err != nil {
		return &PostgresStorage{}, err
	}

	return &PostgresStorage{
		db:      db,
		connStr: conString,
		log:     log,
	}, nil
}

func (PostgresStorage) mapCommonToDBMetric(metric models.Metric) dbMetric {
	return dbMetric{
		MType: metric.MType,
		Name:  metric.ID,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (PostgresStorage) mapDBToCommonMetric(metric dbMetric) models.Metric {
	return models.Metric{
		ID:    metric.Name,
		MType: metric.MType,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (PostgresStorage) retrier(cb func() error) error {

	for i := 0; i < 4; i++ {
		err := cb()
		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != pgerrcode.UniqueViolation {
			return err
		}

		if i < 3 {
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

	return s.retrier(func() error {
		tx, err := s.db.Begin(ctx)

		if err != nil {
			return err
		}

		updStmt := `UPDATE metric SET 
                  metric_type_id = (SELECT id FROM metric_type WHERE metric_type = $1), 
                  metric_name = $2, 
                  value=$3,
                  delta = delta + $4
              	WHERE metric_name = $2;`

		exec, err := tx.Exec(ctx, updStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
		if err != nil {
			return err
		}
		if exec.RowsAffected() == 0 {
			sqlStmt := `INSERT INTO metric (metric_type_id, metric_name, value, delta)
						VALUES ((SELECT id FROM metric_type WHERE metric_type = $1), $2, $3, $4);`

			_, err := tx.Exec(ctx, sqlStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
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

	err := s.retrier(func() error {
		rows, err := s.db.Query(
			ctx,
			`SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m 
    			JOIN metric_type AS t ON m.metric_type_id = t.id WHERE m.metric_name = $1;`, metricName,
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

	err := s.retrier(func() error {
		rows, err := s.db.Query(
			ctx,
			`SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m 
    		JOIN metric_type AS t ON m.metric_type_id = t.id;`,
		)
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

	return s.retrier(func() error {
		updStmt := `UPDATE metric SET 
                  metric_type_id = (SELECT id FROM metric_type WHERE metric_type = $1), 
                  metric_name = $2, 
                  value=$3,
                  delta = delta + $4
              	WHERE metric_name = $2;`

		tx, err := s.db.Begin(ctx)
		if err != nil {
			return err
		}

		batchUpdate := pgx.Batch{}
		for _, metric := range metrics {
			batchUpdate.Queue(updStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
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
			insertStmt := `INSERT INTO metric (metric_type_id, metric_name, value, delta)
							VALUES ((SELECT id FROM metric_type WHERE metric_type = $1), $2, $3, $4);`

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
