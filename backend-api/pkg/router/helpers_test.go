package router

import (
	"backend-api/pkg/terrors"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRenderJSON_Success(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "test-request-id")
	payload := map[string]string{"message": "hello"}
	expectedStatusCode := http.StatusOK

	RenderJSON(ctx, recorder, expectedStatusCode, payload)

	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "test-request-id", recorder.Header().Get(middleware.RequestIDHeader))

	var responseBody map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, payload["message"], responseBody["message"])
}

func TestRenderJSON_MarshalError(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.Background()
	// Functions cannot be marshalled to JSON, this will cause an error
	payload := func() {}
	expectedStatusCode := http.StatusInternalServerError

	RenderJSON(ctx, recorder, http.StatusOK, payload) // StatusOK is intentional to test override

	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.True(t, len(recorder.Body.String()) > 0)
}

func TestRenderError_WithTerror(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "test-request-id")
	terror := terrors.RecordNotFound("resource not found")
	expectedStatusCode := http.StatusBadRequest

	RenderError(ctx, recorder, terror)

	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "test-request-id", recorder.Header().Get(middleware.RequestIDHeader))
}

func TestRenderError_WithGenericError(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "another-id")
	genericErr := errors.New("a generic error occurred")
	expectedStatusCode := http.StatusInternalServerError // Default for unknown errors

	RenderError(ctx, recorder, genericErr)

	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "another-id", recorder.Header().Get(middleware.RequestIDHeader))
}

func TestDefaultSuccessResponseBody(t *testing.T) {
	expected := map[string]string{
		"code":    "Success",
		"message": "action completed successfully",
	}
	result := DefaultSuccessResponseBody()
	assert.Equal(t, expected, result)
}
