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

func TestSendOTP_AntiEnumeration(t *testing.T) {
	r, controller := setupTestApp(t)

	body := []byte(`{"email":"unknown@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ctx := r.Mux.NewContext(req, rec)
	controller.handleSendOTP(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSendOTP_CooldownViolation(t *testing.T) {
	r, controller := setupTestApp(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// First request succeeds
	body := []byte(`{"email":"admin@example.com"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	ctx := r.Mux.NewContext(req1, rec1)
	controller.handleSendOTP(ctx)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Immediate second request triggers cooldown (200 OK but does nothing)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	ctx = r.Mux.NewContext(req2, rec2)
	controller.handleSendOTP(ctx)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestVerifyOTP_InvalidCode_IncrementsAttempts(t *testing.T) {
	r, controller := setupTestApp(t)

	// Invalidate previous verification and create a fresh one
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	bodySend := []byte(`{"email":"admin@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()

	ctx := r.Mux.NewContext(reqSend, recSend)
	controller.handleSendOTP(ctx)
	assert.Equal(t, http.StatusOK, recSend.Code)

	// Submit wrong code
	bodyVerify := []byte(`{"email":"admin@example.com","code":"000000"}`)
	reqVerify := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	recVerify := httptest.NewRecorder()

	ctx = r.Mux.NewContext(reqVerify, recVerify)
	controller.handleVerifyOTP(ctx)
	assert.Equal(t, http.StatusBadRequest, recVerify.Code)

	// Check attempts incremented to 1
	var attempts int
	err = testDB.NewSelect().Table("verifications").Column("attempts").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Scan(t.Context(), &attempts)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestVerifyOTP_BruteForceDefense(t *testing.T) {
	r, controller := setupTestApp(t)

	// Clean verifications
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	bodySend := []byte(`{"email":"admin@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()

	ctx := r.Mux.NewContext(reqSend, recSend)
	controller.handleSendOTP(ctx)
	assert.Equal(t, http.StatusOK, recSend.Code)

	// Attempt 1: wrong code -> attempts=1
	bodyVerify := []byte(`{"email":"admin@example.com","code":"000001"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	ctx = r.Mux.NewContext(req1, rec1)
	controller.handleVerifyOTP(ctx)
	assert.Equal(t, http.StatusBadRequest, rec1.Code)

	// Attempt 2: wrong code -> attempts=2
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	ctx = r.Mux.NewContext(req2, rec2)
	controller.handleVerifyOTP(ctx)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)

	// Attempt 3: wrong code -> reached 3, record deleted
	req3 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()

	ctx = r.Mux.NewContext(req3, rec3)
	controller.handleVerifyOTP(ctx)
	assert.Equal(t, http.StatusBadRequest, rec3.Code)

	// Verification record should now be deleted
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
