package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"employees-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("запись не найдена")
	ErrDuplicatePhone  = errors.New("телефон уже существует")
)

type contextKey string

const dbTimeKey contextKey = "db_time_ms"

type EmployeeRepository struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

func setDBTime(ctx context.Context, duration time.Duration) {
	if v := ctx.Value(dbTimeKey); v != nil {
		if ptr, ok := v.(*int64); ok {
			*ptr = duration.Milliseconds()
		}
	}
}

func GetDBTime(ctx context.Context) int64 {
	if v := ctx.Value(dbTimeKey); v != nil {
		if ptr, ok := v.(*int64); ok {
			return *ptr
		}
	}
	return 0
}

func WithDBTime(ctx context.Context) (context.Context, *int64) {
	var dbTime int64
	return context.WithValue(ctx, dbTimeKey, &dbTime), &dbTime
}

func (r *EmployeeRepository) Create(ctx context.Context, req domain.CreateEmployeeRequest) (*domain.Employee, error) {
	query := `
		INSERT INTO employees (full_name, phone, city)
		VALUES ($1, $2, $3)
		RETURNING id, full_name, phone, city, created_at, updated_at
	`

	start := time.Now()
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, req.FullName, req.Phone, req.City).Scan(
		&emp.ID,
		&emp.FullName,
		&emp.Phone,
		&emp.City,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	setDBTime(ctx, time.Since(start))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicatePhone
		}
		return nil, fmt.Errorf("ошибка создания сотрудника: %w", err)
	}

	return &emp, nil
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	query := `
		SELECT id, full_name, phone, city, created_at, updated_at
		FROM employees
		WHERE id = $1
	`

	start := time.Now()
	var emp domain.Employee
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&emp.ID,
		&emp.FullName,
		&emp.Phone,
		&emp.City,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	setDBTime(ctx, time.Since(start))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ошибка получения сотрудника: %w", err)
	}

	return &emp, nil
}

func (r *EmployeeRepository) HealthCheck(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
