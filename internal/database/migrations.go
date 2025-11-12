package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("создание таблицы миграций: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("чтение директории миграций: %w", err)
	}

	var upMigrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			upMigrations = append(upMigrations, file.Name())
		}
	}
	sort.Strings(upMigrations)

	for _, filename := range upMigrations {
		name := strings.TrimSuffix(filename, ".up.sql")

		applied, err := isMigrationApplied(ctx, pool, name)
		if err != nil {
			return fmt.Errorf("проверка миграции %s: %w", name, err)
		}

		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("чтение миграции %s: %w", filename, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("выполнение миграции %s: %w", filename, err)
		}

		if err := markMigrationApplied(ctx, pool, name); err != nil {
			return fmt.Errorf("сохранение миграции %s: %w", name, err)
		}
	}

	return nil
}

func createMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`
	_, err := pool.Exec(ctx, query)
	return err
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)"
	err := pool.QueryRow(ctx, query, name).Scan(&exists)
	return exists, err
}

func markMigrationApplied(ctx context.Context, pool *pgxpool.Pool, name string) error {
	query := "INSERT INTO schema_migrations (name) VALUES ($1)"
	_, err := pool.Exec(ctx, query, name)
	return err
}
