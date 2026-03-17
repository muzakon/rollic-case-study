package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"server/internal/config"
	"server/internal/server"

	_ "server/migrations"

	"github.com/gofiber/fiber/v3"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Shared state available to all test files in this package.
var (
	testApp *fiber.App
	testDB  *gorm.DB
	testCfg *config.Config
)

// testDBName is the name of the ephemeral database created for this test run.
var testDBName string

// TestMain is the entry point for all E2E tests.
// It creates a temporary database, runs goose migrations, builds the Fiber app,
// executes all tests, and tears everything down.
func TestMain(m *testing.M) {
	cfg := testConfig()
	testCfg = cfg

	// Silence goose migration logs to keep test output clean
	goose.SetLogger(goose.NopLogger())

	// 1. Create temporary database
	if err := createTestDB(cfg); err != nil {
		log.Fatalf("failed to create test database: %v", err)
	}

	// 2. Connect GORM to the test database
	gormDB, err := connectGORM(cfg)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}
	testDB = gormDB

	// 3. Run goose migrations
	if err := runMigrations(gormDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// 4. Build Fiber app with logging disabled to keep test output clean
	testLogger := zerolog.Nop()
	testApp = server.New(cfg, &testLogger, gormDB)

	// 5. Run tests
	code := m.Run()

	// 6. Teardown: close connection, drop database
	sqlDB, _ := gormDB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	if err := dropTestDB(cfg); err != nil {
		log.Printf("warning: failed to drop test database %s: %v", testDBName, err)
	}

	os.Exit(code)
}

// testConfig builds a Config struct with hardcoded values for the test environment.
// It does NOT read .env to avoid interfering with the dev database.
func testConfig() *config.Config {
	testDBName = fmt.Sprintf("test_db_%d", time.Now().UnixNano())

	return &config.Config{
		App: config.AppConfig{
			Name: "leaderboard-api-test",
			Env:  config.EnvTesting,
		},
		Database: config.DatabaseConfig{
			Host:            getEnvOrDefault("TEST_DB_HOST", "postgres"),
			Port:            getEnvOrDefault("TEST_DB_PORT", "5432"),
			User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
			Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
			Name:            testDBName,
			SSLMode:         "disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5,
		},
		Pagination: config.PaginationConfig{
			MaxLimit: 100,
		},
	}
}

// getEnvOrDefault returns the value of the environment variable or the given default.
func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// createTestDB connects to the default "postgres" database and creates the test database.
func createTestDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		return fmt.Errorf("create database %s: %w", testDBName, err)
	}

	return nil
}

// connectGORM opens a GORM connection to the test database.
func connectGORM(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password,
		cfg.Database.Name, cfg.Database.SSLMode,
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// runMigrations uses goose to apply all registered migrations to the test database.
func runMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("extract sql.DB: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.RunContext(context.Background(), "up", sqlDB, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

// dropTestDB connects to the default "postgres" database and drops the test database.
func dropTestDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	// Terminate existing connections to the test database
	_, _ = db.Exec(fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()",
		testDBName,
	))

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return fmt.Errorf("drop database %s: %w", testDBName, err)
	}

	return nil
}
