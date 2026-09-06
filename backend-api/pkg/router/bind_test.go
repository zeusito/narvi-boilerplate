//nolint:noctx
package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestEmptyBody(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	var testStruct TestStruct
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	err := BindBody(req, &testStruct)

	// Verify
	assert.NotNil(t, err)
	assert.Equal(t, "empty request body", err.Error())
}

func TestUnsupportedMediaType(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	var testStruct TestStruct
	body := []byte(`{"field1":"test","field2":1}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	err := BindBody(req, &testStruct)

	// Verify
	assert.NotNil(t, err)
	assert.Equal(t, "unsupported content type", err.Error())
}

func TestBindJSONOk(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	var testStruct TestStruct
	body := []byte(`{"field1":"test","field2":1}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	err := BindBody(req, &testStruct)

	// Verify
	assert.Nil(t, err)
	assert.Equal(t, "test", testStruct.Field1)
	assert.Equal(t, 1, testStruct.Field2)
}

func TestBindJSONInvalid(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	var testStruct TestStruct
	body := []byte(`{"field1":"test","field2":"invalid"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	err := BindBody(req, &testStruct)

	// Verify
	assert.NotNil(t, err)
	assert.Equal(t, "invalid json payload", err.Error())
}

func TestBindJSONValidationFailed(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1" validate:"required"`
		Field2 int    `json:"field2"`
	}

	var testStruct TestStruct
	body := []byte(`{"field2":1}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	err := BindBody(req, &testStruct)

	log.Info().Msg(err.Error())

	// Verify
	assert.NotNil(t, err)
	assert.True(t, len(err.Error()) > 0)
}

func TestValidateStructOk(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1" validate:"required"`
		Field2 int    `json:"field2"`
	}

	testStruct := &TestStruct{
		Field1: "test",
		Field2: 1,
	}
	err := Validate(testStruct)

	// Verify
	assert.Nil(t, err)
}

func TestValidateStructInvalid(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1" validate:"required"`
		Field2 int    `json:"field2"`
	}

	testStruct := &TestStruct{}
	err := Validate(testStruct)

	// Verify
	assert.NotNil(t, err)
	assert.True(t, len(err.Error()) > 0)
}
