package transport

import (
	"context"
	"net/http"
	"time"

	"employees-api/internal/repository"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func (h *Handler) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx, dbTimePtr := repository.WithDBTime(r.Context())

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r.WithContext(ctx))

		duration := time.Since(start)
		requestID, _ := r.Context().Value(requestIDKey).(string)

		logData := map[string]interface{}{
			"ид_запроса":  requestID,
			"метод":       r.Method,
			"путь":        r.URL.Path,
			"статус":      lrw.statusCode,
			"задержка_мс": duration.Milliseconds(),
			"адрес":       r.RemoteAddr,
		}

		if dbTimePtr != nil && *dbTimePtr > 0 {
			logData["время_бд_мс"] = *dbTimePtr
		}

		h.logger.Info("http_запрос", logData)
	})
}

func (h *Handler) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := r.Context().Value(requestIDKey).(string)
				h.logger.Error("восстановление_паники", map[string]interface{}{
					"ид_запроса": requestID,
					"тип_ошибки": "паника",
				})

				respondError(w, ErrorResponse{
					Code:    "internal_error",
					Message: "Внутренняя ошибка сервера",
				}, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
