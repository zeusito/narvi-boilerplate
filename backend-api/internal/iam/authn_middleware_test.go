package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthRateLimiter_EnforcesLimit(t *testing.T) {
	e := echo.New()
	limiter := NewAuthRateLimiter(0.1, 2) // 2 requests allowed burst

	handler := limiter(func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Request 1: OK
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.RemoteAddr = "192.0.2.1:1234"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	err := handler(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Request 2: OK
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.RemoteAddr = "192.0.2.1:1234"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = handler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Request 3: Exceeds burst -> returns 429 error
	req3 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req3.RemoteAddr = "192.0.2.1:1234"
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	err = handler(c3)
	assert.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusTooManyRequests, httpErr.Code)
}
