package repository

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db      *pgxpool.Pool
	connStr string
}

type dbMetric struct {
	ID    int64    `db:"id"`
	MType string   `db:"metric_type"`
	Name  string   `db:"metric_name"`
	Value *float64 `db:"value"`
	Delta *int64   `db:"delta"`
}

func NewPostgresStorage(ctx context.Context, conString string) (*PostgresStorage, error) {
	db, err := pgxpool.New(ctx, conString)
	if err != nil {
		return &PostgresStorage{}, err
	}

	return &PostgresStorage{
		db:      db,
		connStr: conString,
	}, nil
}

func (PostgresStorage) mapCommonToDBMetric(metric models.Metrics) dbMetric {
	return dbMetric{
		MType: metric.MType,
		Name:  metric.ID,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (PostgresStorage) mapDBToCommonMetric(metric dbMetric) models.Metrics {
	return models.Metrics{
		ID:    metric.Name,
		MType: metric.MType,
		Value: metric.Value,
		Delta: metric.Delta,
	}
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	if s.db == nil {
		return errors.New("database not initialized")
	}
	return s.db.Ping(ctx)
}

func (s *PostgresStorage) Save(ctx context.Context, metric models.Metrics) error {
	updStmt := `UPDATE metric SET 
                  metric_type_id = (SELECT id FROM metric_type WHERE metric_type = $1), 
                  metric_name = $2, 
                  value=$3,
                  delta = delta + $4
              	WHERE metric_name = $2;`

	exec, err := s.db.Exec(ctx, updStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
	if err != nil {
		return err
	}
	if exec.RowsAffected() == 0 {
		sqlStmt := `INSERT INTO metric (metric_type_id, metric_name, value, delta)
						VALUES ((SELECT id FROM metric_type WHERE metric_type = $1), $2, $3, $4);`

		_, err := s.db.Exec(ctx, sqlStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) Find(ctx context.Context, metric string) (models.Metrics, error) {
	rows, err := s.db.Query(ctx, `SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m 
    	JOIN metric_type AS t ON m.metric_type_id = t.id WHERE m.metric_name = $1;`, metric)

	if err != nil {
		return models.Metrics{}, err
	}

	defer rows.Close()

	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbMetric])
	if err != nil {
		return models.Metrics{}, err
	}

	return s.mapDBToCommonMetric(m), nil
}

func (s *PostgresStorage) GetAll(ctx context.Context) (map[string]models.Metrics, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m 
    		JOIN metric_type AS t ON m.metric_type_id = t.id;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbMetric])
	if err != nil {
		return nil, err
	}
	cMetrics := make(map[string]models.Metrics, len(metrics))
	for _, m := range metrics {
		cMetrics[m.Name] = s.mapDBToCommonMetric(m)
	}
	return cMetrics, nil
}

func (s *PostgresStorage) SaveAll(ctx context.Context, metrics map[string]models.Metrics) error {

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

	defer batchResult.Close()

	insertRows := make([]models.Metrics, 0, len(metrics))

	for _, metric := range metrics {
		row, err := batchResult.Exec()
		if err != nil {
			insertRows = append(insertRows, metric)
		}
		if row.RowsAffected() == 0 {
			insertRows = append(insertRows, metric)
		}
	}

	if len(insertRows) > 0 {
		insertStmt := `INSERT INTO metric (metric_type_id, metric_name, value, delta)
							VALUES ((SELECT id FROM metric_type WHERE metric_type = $1), $2, $3, $4);`

		insertBatch := pgx.Batch{}

		for _, metric := range insertRows {
			insertBatch.Queue(insertStmt, metric.MType, metric.ID, metric.Value, metric.Delta)
		}

		batchResult = tx.SendBatch(ctx, &insertBatch)
		defer batchResult.Close()

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

	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) Close() {

	s.db.Close()
}
