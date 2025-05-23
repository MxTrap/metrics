package migrator

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	connStr, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err, "failed to get connection string")

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "failed to create pgx pool")

	cleanup := func() {
		_ = pgContainer.Terminate(ctx)
	}
	return pool, cleanup
}

func setupTestMigrations(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err, "failed to create temp migrations dir")

	upMigration := `
CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);
`
	downMigration := `
DROP TABLE test_table;
`
	err = os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte(upMigration), 0644)
	require.NoError(t, err, "failed to write up migration")
	err = os.WriteFile(filepath.Join(dir, "0001_init.down.sql"), []byte(downMigration), 0644)
	require.NoError(t, err, "failed to write down migration")

	cleanup := func() {
		os.RemoveAll(dir)
	}
	return dir, cleanup
}

func TestNewMigrator(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	migrationsDir, migrationsCleanup := setupTestMigrations(t)
	defer migrationsCleanup()

	migrator, err := NewMigrator(pool, migrationsDir)
	require.NoError(t, err, "failed to create migrator")
	assert.NotNil(t, migrator, "migrator should not be nil")
	assert.NotNil(t, migrator.migrator, "migrator.migrator should not be nil")
	assert.NotNil(t, migrator.db, "migrator.db should not be nil")
}

func TestNewMigratorInvalidMigrationsPath(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	invalidPath := "/invalid/path"
	_, err := NewMigrator(pool, invalidPath)
	assert.Error(t, err, "should fail with invalid migrations path")
}

func TestInitializeDB(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	migrationsDir, migrationsCleanup := setupTestMigrations(t)
	defer migrationsCleanup()

	migrator, err := NewMigrator(pool, migrationsDir)
	require.NoError(t, err, "failed to create migrator")

	err = migrator.InitializeDB()
	assert.NoError(t, err, "initialize DB should succeed")

	ctx := context.Background()
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'test_table'
		)
	`).Scan(&exists)
	require.NoError(t, err, "failed to check table existence")
	assert.True(t, exists, "test_table should exist after migration")
}

func TestMigratorClose(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	migrationsDir, migrationsCleanup := setupTestMigrations(t)
	defer migrationsCleanup()

	migrator, err := NewMigrator(pool, migrationsDir)
	require.NoError(t, err, "failed to create migrator")

	err = migrator.InitializeDB()
	assert.NoError(t, err, "initialize DB should succeed")

	err = migrator.db.Ping()
	assert.Error(t, err, "db should be closed after InitializeDB")
}

func TestInitializeDBNoChange(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	migrationsDir, migrationsCleanup := setupTestMigrations(t)
	defer migrationsCleanup()

	migrator, err := NewMigrator(pool, migrationsDir)
	require.NoError(t, err, "failed to create migrator")

	err = migrator.InitializeDB()
	assert.NoError(t, err, "first initialize DB should succeed")

	migrator2, err := NewMigrator(pool, migrationsDir)
	require.NoError(t, err, "failed to create second migrator")
	err = migrator2.InitializeDB()
	assert.NoError(t, err, "second initialize DB should succeed (ErrNoChange)")
}

func TestInitializeDBInvalidMigration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	dir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err, "failed to create temp migrations dir")
	defer os.RemoveAll(dir)

	invalidMigration := `INVALID SQL;`
	err = os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte(invalidMigration), 0644)
	require.NoError(t, err, "failed to write invalid migration")

	migrator, err := NewMigrator(pool, dir)
	require.NoError(t, err, "failed to create migrator")

	err = migrator.InitializeDB()
	assert.Error(t, err, "initialize DB should fail with invalid migration")
}
