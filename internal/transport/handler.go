package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"employees-api/internal/domain"
	"employees-api/internal/repository"
	"employees-api/internal/service"

	"github.com/google/uuid"
)

type Handler struct {
	service *service.EmployeeService
	logger  *Logger
}

func NewHandler(svc *service.EmployeeService, logger *Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/employees", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateEmployee(w, r)
		} else {
			respondError(w, ErrorResponse{
				Code:    "method_not_allowed",
				Message: "Метод не поддерживается",
			}, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/employees/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetEmployee(w, r)
		} else {
			respondError(w, ErrorResponse{
				Code:    "method_not_allowed",
				Message: "Метод не поддерживается",
			}, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/healthz", h.HealthCheck)

	handler := h.requestIDMiddleware(mux)
	handler = h.loggingMiddleware(handler)
	handler = h.recoverMiddleware(handler)

	return handler
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if r.Header.Get("Content-Type") != "application/json" {
		respondError(w, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type должен быть application/json",
		}, http.StatusBadRequest)
		return
	}

	var req domain.CreateEmployeeRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, 1024*1024))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		respondError(w, ErrorResponse{
			Code:    "invalid_json",
			Message: "Невалидный JSON",
		}, http.StatusBadRequest)
		return
	}

	emp, err := h.service.CreateEmployee(ctx, req)
	if err != nil {
		var validationErr *service.ValidationErrors
		if errors.As(err, &validationErr) {
			details := make(map[string]interface{})
			for _, e := range validationErr.Errors {
				details[e.Field] = e.Message
			}
			respondError(w, ErrorResponse{
				Code:    "validation_error",
				Message: "Ошибка валидации",
				Details: details,
			}, http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, repository.ErrDuplicatePhone) {
			respondError(w, ErrorResponse{
				Code:    "duplicate_phone",
				Message: "Телефон уже существует",
			}, http.StatusConflict)
			return
		}

		h.logger.Error("ошибка_создания_сотрудника", map[string]interface{}{
			"тип_ошибки": "внутренняя",
		})
		respondError(w, ErrorResponse{
			Code:    "internal_error",
			Message: "Внутренняя ошибка сервера",
		}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/v1/employees/"+emp.ID.String())
	respondJSON(w, emp, http.StatusCreated)
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	idStr := r.URL.Path[len("/v1/employees/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, ErrorResponse{
			Code:    "invalid_id",
			Message: "Невалидный ID",
		}, http.StatusBadRequest)
		return
	}

	emp, err := h.service.GetEmployeeByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, ErrorResponse{
				Code:    "not_found",
				Message: "Сотрудник не найден",
			}, http.StatusNotFound)
			return
		}

		h.logger.Error("ошибка_получения_сотрудника", map[string]interface{}{
			"тип_ошибки": "внутренняя",
		})
		respondError(w, ErrorResponse{
			Code:    "internal_error",
			Message: "Внутренняя ошибка сервера",
		}, http.StatusInternalServerError)
		return
	}

	respondJSON(w, emp, http.StatusOK)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.service.HealthCheck(ctx); err != nil {
		h.logger.Error("проверка_здоровья_провалена", map[string]interface{}{
			"тип_ошибки": "база_данных_недоступна",
		})
		respondError(w, ErrorResponse{
			Code:    "unhealthy",
			Message: "Сервис недоступен",
		}, http.StatusServiceUnavailable)
		return
	}

	respondJSON(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err ErrorResponse, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}
