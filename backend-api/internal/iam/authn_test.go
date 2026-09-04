package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/router"
	"backend-api/pkg/toolbox/hasher"
	"backend-api/pkg/toolbox/testbox"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

var (
	testDB     *bun.DB
	testHasher hasher.Hasher
	fakeMailer mailer.Mailer
)

const testSecret = "dGVzdC1zZWNyZXQta2V5LTMyLWJ5dGVzLWxvbmctISE="

func TestMain(m *testing.M) {
	ctx := context.Background()

	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	schemaPath := filepath.Join(rootDir, "db", "schema.sql")
	testDataPath := filepath.Join(rootDir, "db", "testdata", "iam.sql")

	conn, cleanup, err := testbox.InitPostgresqlContainer(ctx, []string{
		schemaPath,
		testDataPath,
	})
	if err != nil {
		panic("failed to initialize test container: " + err.Error())
	}
	defer cleanup()

	testDB = conn

	h, err := hasher.NewHmacSHA256(testSecret)
	if err != nil {
		panic("failed to initialize test hasher: " + err.Error())
	}
	testHasher = h

	fakeMailer = mailer.NewFakeMailer()

	os.Exit(m.Run())
}

func setupTestApp(t *testing.T) (*router.HttpRouter, *authnController) {
	r := router.NewRouter()
	identityRepo := newIdentityRepository(testDB)
	orgRepo := newOrganizationRepository(testDB)
	verificationRepo := newVerificationRepository(testDB)
	sessionRepo := newSessionRepository(testDB)

	useCases := newAuthnUseCases(orgRepo, identityRepo, verificationRepo, sessionRepo, fakeMailer, testHasher)
	sessionManager := newSessionManager(useCases)
	controller := newAuthnController(r.Mux, sessionManager, useCases)

	return r, controller
}

func TestSendOTP_ValidUser(t *testing.T) {
	r, controller := setupTestApp(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	body := []byte(`{"email":"admin@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ctx := r.Mux.NewContext(req, rec)

	controller.handleSendOTP(ctx)

	// Verify database record created
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
