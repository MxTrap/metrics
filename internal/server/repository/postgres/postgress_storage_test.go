package postgres

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

const dbName = "postgres"
const username = "postgres"
const password = "postgres"

func createTestContainer() (*postgres.PostgresContainer, error) {
	ctx := context.Background()

	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}

type cleanupFn func(ctx context.Context)

func createPool() (*pgxpool.Pool, cleanupFn, error) {
	ctx := context.Background()
	ctr, err := createTestContainer()
	if err != nil {
		return nil, nil, err
	}

	connString, err := ctr.ConnectionString(ctx)
	if err != nil {
		return nil, nil, err
	}
	pgPool, err := pgxpool.New(
		ctx,
		connString,
	)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func(ctx context.Context) {
		pgPool.Close()

		fmt.Println(ctr.Terminate(ctx))
	}

	return pgPool, cleanup, nil
}

func setupStorage() (*Storage, cleanupFn, error) {
	pgPool, cleanupFn, err := createPool()
	if err != nil {
		return nil, cleanupFn, err
	}

	log := logger.NewLogger()

	_, err = pgPool.Exec(
		context.Background(),
		"CREATE TABLE metric_type( id SERIAL PRIMARY KEY, metric_type VARCHAR(20) NOT NULL UNIQUE);",
	)
	if err != nil {
		return nil, nil, err
	}

	_, err = pgPool.Exec(
		context.Background(),
		"INSERT INTO metric_type(id, metric_type) VALUES (1, 'gauge'), (2, 'counter');",
	)
	if err != nil {
		return nil, nil, err
	}

	_, err = pgPool.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS metric
		(
			id              SERIAL PRIMARY KEY,
			metric_type_id  INT,
			metric_name     VARCHAR,
			value           DOUBLE PRECISION,
			delta           BIGINT,
			CONSTRAINT fk_metric_metric_type
		FOREIGN KEY (metric_type_id)
		REFERENCES metric_type (id)
		)`,
	)
	if err != nil {
		return nil, nil, err
	}

	return &Storage{db: pgPool, log: log}, cleanupFn, nil
}

func TestNewPostgresStorage(t *testing.T) {
	pool, cleanup, err := createPool()
	defer cleanup(context.Background())

	log := logger.NewLogger()
	storage, err := NewPostgresStorage(pool, log)
	require.NoError(t, err, "failed to create storage")
	assert.NotNil(t, storage, "storage should not be nil")
	assert.Equal(t, pool, storage.db, "db pool should match")
	assert.Equal(t, log, storage.log, "logger should match")
}

func TestPing(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")

	ctx := context.Background()
	err = storage.Ping(ctx)
	assert.NoError(t, err, "ping should succeed")

	storage.Close()
	err = storage.Ping(ctx)
	assert.Error(t, err, "ping should fail on closed pool")
}

func TestSave(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")

	ctx := context.Background()
	metric := models.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Value: new(float64),
	}
	*metric.Value = 42.0

	err = storage.Save(ctx, metric)
	assert.NoError(t, err, "save should succeed")
}

func TestFind(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")

	ctx := context.Background()
	metric := models.Metric{
		ID:    "testCounter",
		MType: "counter",
		Delta: new(int64),
	}
	*metric.Delta = 100

	err = storage.Save(ctx, metric)
	require.NoError(t, err, "failed to save metric")

	found, err := storage.Find(ctx, "testCounter")
	assert.NoError(t, err, "find should succeed")
	assert.Equal(t, metric.ID, found.ID)
	assert.Equal(t, metric.MType, found.MType)
	assert.Equal(t, *metric.Delta, *found.Delta)

	_, err = storage.Find(ctx, "nonexistent")
	assert.Error(t, err, "find should fail for nonexistent metric")
}

func TestGetAll(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")
	require.NoError(t, err, "failed to create storage")

	ctx := context.Background()
	metrics := map[string]models.Metric{
		"gauge1": {
			ID:    "gauge1",
			MType: "gauge",
			Value: new(float64),
		},
		"counter1": {
			ID:    "counter1",
			MType: "counter",
			Delta: new(int64),
		},
	}
	*metrics["gauge1"].Value = 42.0
	*metrics["counter1"].Delta = 100

	err = storage.SaveAll(ctx, metrics)
	require.NoError(t, err, "failed to save metrics")

	result, err := storage.GetAll(ctx)
	assert.NoError(t, err, "get all should succeed")
	assert.Len(t, result, 2, "should return 2 metrics")
	assert.Equal(t, metrics["gauge1"], result["gauge1"])
	assert.Equal(t, metrics["counter1"], result["counter1"])
}

func TestSaveAll(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")

	ctx := context.Background()
	metrics := map[string]models.Metric{
		"gauge1": {
			ID:    "gauge1",
			MType: "gauge",
			Value: new(float64),
		},
		"counter1": {
			ID:    "counter1",
			MType: "counter",
			Delta: new(int64),
		},
	}
	*metrics["gauge1"].Value = 42.0
	*metrics["counter1"].Delta = 100

	err = storage.SaveAll(ctx, metrics)
	assert.NoError(t, err, "save all should succeed")

}

func TestClose(t *testing.T) {
	storage, cleanup, err := setupStorage()
	defer cleanup(context.Background())

	require.NoError(t, err, "failed to create storage")

	storage.Close()
	err = storage.Ping(context.Background())
	assert.Error(t, err, "ping should fail after close")
}
