package terrors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreconditionFailed(t *testing.T) {
	err := PreconditionFailed("test")
	assert.True(t, err.ErrMessage == "test")
	assert.True(t, err.ErrCode == "PreconditionFailed")
	assert.True(t, err.HttpStatusCode == http.StatusBadRequest)
}

func TestForbidden(t *testing.T) {
	err := Forbidden("test")
	assert.True(t, err.ErrMessage == "test")
	assert.True(t, err.ErrCode == "ActionForbidden")
	assert.True(t, err.HttpStatusCode == http.StatusForbidden)
}

func TestRecordNotFound(t *testing.T) {
	err := RecordNotFound("test")
	assert.True(t, err.ErrMessage == "test")
	assert.True(t, err.ErrCode == "RecordNotFound")
	assert.True(t, err.HttpStatusCode == http.StatusBadRequest)
}

func TestUnknown(t *testing.T) {
	err := Unknown("test")
	assert.True(t, err.ErrMessage == "test")
	assert.True(t, err.ErrCode == "UnknownError")
	assert.True(t, err.HttpStatusCode == http.StatusInternalServerError)
}

func TestTypeAssertion(t *testing.T) {
	var err error = PreconditionFailed("test")

	terr, ok := err.(*Terror)

	assert.True(t, ok)
	assert.True(t, terr.ErrMessage == "test")
	assert.True(t, terr.ErrCode == "PreconditionFailed")
}
