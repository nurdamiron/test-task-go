package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"employees-api/internal/config"
	"employees-api/internal/database"
	"employees-api/internal/domain"
	"employees-api/internal/repository"
	"employees-api/internal/service"
	"employees-api/internal/transport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testServer struct {
	baseURL string
	cleanup func()
}

func setupTestServer(t *testing.T) *testServer {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())

	cfg := &config.Config{
		Port:                "8888",
		PostgresDSN:         dsn,
		DBMaxConns:          10,
		DBMinConns:          2,
		DBMaxConnLifetime:   time.Hour,
		DBHealthCheckPeriod: time.Minute,
		ReadTimeout:         5 * time.Second,
		WriteTimeout:        10 * time.Second,
		RunMigrations:       true,
	}

	pool, err := database.NewPool(ctx, cfg)
	require.NoError(t, err)

	err = database.RunMigrations(ctx, pool, "../migrations")
	require.NoError(t, err)

	repo := repository.NewEmployeeRepository(pool)
	svc := service.NewEmployeeService(repo)
	logger := transport.NewLogger()
	handler := transport.NewHandler(svc, logger)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler.Routes(),
	}

	go func() {
		server.ListenAndServe()
	}()

	time.Sleep(500 * time.Millisecond)

	return &testServer{
		baseURL: "http://localhost:" + cfg.Port,
		cleanup: func() {
			server.Shutdown(ctx)
			pool.Close()
			pgContainer.Terminate(ctx)
		},
	}
}

func TestCreateEmployee_Success(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	reqBody := domain.CreateEmployeeRequest{
		FullName: "Иван Иванов",
		Phone:    "+79991234567",
		City:     "Москва",
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.baseURL+"/v1/employees", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var employee domain.Employee
	err = json.NewDecoder(resp.Body).Decode(&employee)
	require.NoError(t, err)

	assert.NotEmpty(t, employee.ID)
	assert.Equal(t, "Иван Иванов", employee.FullName)
	assert.Equal(t, "+79991234567", employee.Phone)
	assert.Equal(t, "Москва", employee.City)
}

func TestCreateEmployee_ValidationError(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	reqBody := domain.CreateEmployeeRequest{
		FullName: "A",
		Phone:    "invalid",
		City:     "M",
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.baseURL+"/v1/employees", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	var errResp transport.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "validation_error", errResp.Code)
	assert.NotEmpty(t, errResp.Details)
}

func TestCreateEmployee_DuplicatePhone(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	reqBody := domain.CreateEmployeeRequest{
		FullName: "Петр Петров",
		Phone:    "+79991111111",
		City:     "Санкт-Петербург",
	}

	body, _ := json.Marshal(reqBody)
	resp, _ := http.Post(srv.baseURL+"/v1/employees", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()

	resp, err := http.Post(srv.baseURL+"/v1/employees", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetEmployee_Success(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	reqBody := domain.CreateEmployeeRequest{
		FullName: "Сергей Сергеев",
		Phone:    "+79992222222",
		City:     "Казань",
	}

	body, _ := json.Marshal(reqBody)
	createResp, _ := http.Post(srv.baseURL+"/v1/employees", "application/json", bytes.NewBuffer(body))

	var created domain.Employee
	json.NewDecoder(createResp.Body).Decode(&created)
	createResp.Body.Close()

	resp, err := http.Get(srv.baseURL + "/v1/employees/" + created.ID.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var employee domain.Employee
	err = json.NewDecoder(resp.Body).Decode(&employee)
	require.NoError(t, err)

	assert.Equal(t, created.ID, employee.ID)
	assert.Equal(t, "Сергей Сергеев", employee.FullName)
}

func TestGetEmployee_NotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	resp, err := http.Get(srv.baseURL + "/v1/employees/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHealthCheck(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.cleanup()

	resp, err := http.Get(srv.baseURL + "/v1/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "ok", result["status"])
}
