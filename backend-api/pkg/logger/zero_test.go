package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMiddleware_InjectsTraceID(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Logger()
	zerolog.DefaultContextLogger = &log.Logger

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	expectedTraceID := "trace-12345"
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, expectedTraceID)
	req = req.WithContext(ctx)

	handlerExecuted := false
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecuted = true
		log.Ctx(r.Context()).Info().Msg("test message inside handler")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !handlerExecuted {
		t.Fatal("handler was not executed")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v, raw: %s", err, buf.String())
	}

	if traceID, ok := logEntry["traceId"].(string); !ok || traceID != expectedTraceID {
		t.Fatalf("expected traceId %q, got %v", expectedTraceID, logEntry["traceId"])
	}

	if msg, ok := logEntry["message"].(string); !ok || msg != "test message inside handler" {
		t.Fatalf("expected message 'test message inside handler', got %v", logEntry["message"])
	}
}

func TestMiddleware_NoTraceID(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Logger()
	zerolog.DefaultContextLogger = &log.Logger

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Ctx(r.Context()).Info().Msg("no trace id log")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v, raw: %s", err, buf.String())
	}

	if _, ok := logEntry["traceId"]; ok {
		t.Fatalf("expected no traceId key when request ID is empty, got %v", logEntry["traceId"])
	}
}
