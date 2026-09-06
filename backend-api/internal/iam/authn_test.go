package iam

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"
	"backend-api/pkg/toolbox/testbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

var (
	testDB     *bun.DB
	testHasher hasher.Hasher
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

	testDB = conn

	h, err := hasher.NewHmacSHA256(testSecret)
	if err != nil {
		panic("failed to initialize test hasher: " + err.Error())
	}
	testHasher = h

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func setupAuthService(t *testing.T) (authService, *mailer.SpyMailer) {
	identityRepo := newIdentityRepository(testDB)
	orgRepo := newOrganizationRepository(testDB)
	verificationRepo := newVerificationRepository(testDB)
	sessionRepo := newSessionRepository(testDB)
	spyMail := mailer.NewSpyMailer()

	return newAuthnService(orgRepo, identityRepo, verificationRepo, sessionRepo, spyMail, testHasher), spyMail
}

func TestSendOTP_ValidUser(t *testing.T) {
	service, _ := setupAuthService(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	// Verify database record created
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSendOTP_AntiEnumeration(t *testing.T) {
	service, _ := setupAuthService(t)

	err := service.SendOTP(t.Context(), &SendOTPRequest{Email: "unknown@example.com"})
	assert.NoError(t, err)
}

func TestSendOTP_CooldownViolation(t *testing.T) {
	service, _ := setupAuthService(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// First request succeeds
	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	// Immediate second request triggers cooldown (returns nil without error, no duplicate created)
	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestVerifyOTP_InvalidCode_IncrementsAttempts(t *testing.T) {
	service, _ := setupAuthService(t)

	// Invalidate previous verification and create a fresh one
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	// Submit wrong code
	resp, err := service.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     "admin@example.com",
		Code:      "000000",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check attempts incremented to 1
	var attempts int
	err = testDB.NewSelect().Table("verifications").Column("attempts").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Scan(t.Context(), &attempts)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestVerifyOTP_BruteForceDefense(t *testing.T) {
	service, _ := setupAuthService(t)

	// Clean verifications
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	// Attempt 1: wrong code -> attempts=1
	_, err = service.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     "admin@example.com",
		Code:      "000001",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	assert.Error(t, err)

	// Attempt 2: wrong code -> attempts=2
	_, err = service.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     "admin@example.com",
		Code:      "000002",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	assert.Error(t, err)

	// Attempt 3: wrong code -> reached 3, record deleted
	_, err = service.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     "admin@example.com",
		Code:      "000003",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	assert.Error(t, err)

	// Verification record should now be deleted
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestVerifyOTP_Success(t *testing.T) {
	service, spy := setupAuthService(t)

	// Clean verifications and sessions for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	_, err = testDB.NewDelete().Table("identity_sessions").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	err = service.SendOTP(t.Context(), &SendOTPRequest{Email: "admin@example.com"})
	require.NoError(t, err)

	code := spy.LastCode()
	require.NotEmpty(t, code)

	// Submit correct code
	resp, err := service.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     "admin@example.com",
		Code:      code,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, 86400, resp.ExpiresIn)

	// Verification record should now be deleted
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Session record should exist in database
	sCount, err := testDB.NewSelect().Table("identity_sessions").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, sCount, 1)
}
