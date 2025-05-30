package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// setupTestDB creates a new test database connection and runs migrations
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Test database configuration
	config := struct {
		host     string
		port     string
		user     string
		password string
		dbname   string
	}{
		host:     getEnv("TEST_DB_HOST", "localhost"),
		port:     getEnv("TEST_DB_PORT", "5433"),
		user:     getEnv("TEST_DB_USER", "postgres"),
		password: getEnv("TEST_DB_PASSWORD", "postgres"),
		dbname:   getEnv("TEST_DB_NAME", "testdb"),
	}

	// Build connection strings
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.host, config.port, config.user, config.password, config.dbname)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.user, config.password, config.host, config.port, config.dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Wait for database to be ready (with retry)
	var maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if err != nil {
		t.Fatalf("Failed to ping database after %d attempts: %v", maxAttempts, err)
	}

	// Run migrations
	if err := runMigrations(dbURL); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db, func() {
		// Clean up test data
		truncateTables(t, db)
		db.Close()
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func runMigrations(dbURL string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Find the project root by looking for the go.mod file
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return fmt.Errorf("could not find project root (go.mod)")
		}
		cwd = parent
	}

	migrationPath := filepath.Join(cwd, "db/migrations")
	m, err := migrate.New(
		"file://"+filepath.ToSlash(migrationPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// truncateTables truncates all tables in the test database
func truncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{"follows", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}

		// Try to reset the sequence for each table
		seqName := fmt.Sprintf("%s_id_seq", table)
		sql := `DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.sequences WHERE sequence_name = '` + seqName + `') THEN
				EXECUTE format('ALTER SEQUENCE %I RESTART WITH 1', '` + seqName + `');
			END IF;
		END
		$$;`
		_, err = db.Exec(sql)
		if err != nil {
			t.Logf("Failed to reset sequence %s: %v", seqName, err)
		}
	}
}
