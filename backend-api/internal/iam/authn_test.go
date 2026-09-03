package iam_test

import (
	"backend-api/internal/iam"
	"backend-api/pkg/router"
	"backend-api/pkg/toolbox/hasher"
	"backend-api/pkg/toolbox/testbox"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

var (
	testDB     *bun.DB
	testHasher hasher.Hasher
)

const testSecret = "dGVzdC1zZWNyZXQta2V5LTMyLWJ5dGVzLWxvbmctISE="

type capturingMailer struct {
	mu        sync.Mutex
	lastEmail string
	lastCode  string
}

func (m *capturingMailer) SendOTPCode(ctx context.Context, email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastEmail = email
	m.lastCode = code
	return nil
}

func (m *capturingMailer) SendInvitation(ctx context.Context, email, kind string) error {
	return nil
}

func (m *capturingMailer) LastCode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCode
}

func (m *capturingMailer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastEmail = ""
	m.lastCode = ""
}

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

	os.Exit(m.Run())
}

func setupTestApp(t *testing.T) (*router.HttpRouter, *capturingMailer, *iam.Module) {
	r := router.NewRouter()
	mail := &capturingMailer{}
	mod := iam.NewModule(r.Mux, testDB, mail, testHasher)
	return r, mail, mod
}

func TestSendOTP_ValidUser(t *testing.T) {
	r, mail, _ := setupTestApp(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	body := []byte(`{"email":"admin@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.Mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp iam.SendOTPResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Sent)

	// Verify mailer received 6-digit code
	code := mail.LastCode()
	assert.Len(t, code, 6)

	// Verify database record created
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSendOTP_AntiEnumeration(t *testing.T) {
	r, mail, _ := setupTestApp(t)

	body := []byte(`{"email":"unknown@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.Mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp iam.SendOTPResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Sent)

	// Mailer should NOT be called
	assert.Empty(t, mail.LastCode())
}

func TestSendOTP_CooldownViolation(t *testing.T) {
	r, _, _ := setupTestApp(t)

	// Clean any previous verifications for admin
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// First request succeeds
	body := []byte(`{"email":"admin@example.com"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Immediate second request triggers cooldown (429)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestVerifyOTP_InvalidCode_IncrementsAttempts(t *testing.T) {
	r, _, _ := setupTestApp(t)

	// Invalidate previous verification and create a fresh one
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	bodySend := []byte(`{"email":"admin@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()
	r.Mux.ServeHTTP(recSend, reqSend)
	require.Equal(t, http.StatusOK, recSend.Code)

	// Submit wrong code
	bodyVerify := []byte(`{"email":"admin@example.com","code":"000000"}`)
	reqVerify := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	reqVerify.Header.Set("Content-Type", "application/json")
	recVerify := httptest.NewRecorder()
	r.Mux.ServeHTTP(recVerify, reqVerify)

	assert.Equal(t, http.StatusUnauthorized, recVerify.Code)

	// Check attempts incremented to 1
	var attempts int
	err = testDB.NewSelect().Table("verifications").Column("attempts").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Scan(t.Context(), &attempts)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestVerifyOTP_BruteForceDefense(t *testing.T) {
	r, _, _ := setupTestApp(t)

	// Clean verifications
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	bodySend := []byte(`{"email":"admin@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()
	r.Mux.ServeHTTP(recSend, reqSend)
	require.Equal(t, http.StatusOK, recSend.Code)

	// Attempt 1: wrong code -> attempts=1
	bodyVerify := []byte(`{"email":"admin@example.com","code":"000001"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusUnauthorized, rec1.Code)

	// Attempt 2: wrong code -> attempts=2
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)

	// Attempt 3: wrong code -> reached 3, record deleted
	req3 := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(bodyVerify))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusUnauthorized, rec3.Code)

	// Verification record should now be deleted
	count, err := testDB.NewSelect().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestVerifyOTP_Success_And_Introspect(t *testing.T) {
	r, mail, _ := setupTestApp(t)

	// Clean verifications
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c403").Exec(t.Context())
	require.NoError(t, err)

	// Send code
	bodySend := []byte(`{"email":"admin@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()
	r.Mux.ServeHTTP(recSend, reqSend)
	require.Equal(t, http.StatusOK, recSend.Code)

	validCode := mail.LastCode()
	require.NotEmpty(t, validCode)

	// Verify code
	verifyBody, _ := json.Marshal(iam.VerifyOTPRequest{
		Email: "admin@example.com",
		Code:  validCode,
	})
	reqVerify := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(verifyBody))
	reqVerify.Header.Set("Content-Type", "application/json")
	recVerify := httptest.NewRecorder()
	r.Mux.ServeHTTP(recVerify, reqVerify)

	assert.Equal(t, http.StatusOK, recVerify.Code)
	var verifyResp iam.VerifyOTPResponse
	err = json.Unmarshal(recVerify.Body.Bytes(), &verifyResp)
	require.NoError(t, err)

	assert.NotEmpty(t, verifyResp.Token)
	assert.Contains(t, verifyResp.Token, "tok_")
	assert.Equal(t, "01a02086-04a2-75a7-ba24-1616b586c403", verifyResp.Identity.ID)
	assert.Equal(t, "admin@example.com", verifyResp.Identity.Email)
	require.NotNil(t, verifyResp.ActiveOrganization)
	assert.Equal(t, "01a02086-04a2-75a7-ba24-12d5872b8c49", verifyResp.ActiveOrganization.ID)
	assert.Equal(t, "Acme Corp", verifyResp.ActiveOrganization.Name)
	assert.Equal(t, "owner", verifyResp.ActiveOrganization.Role)

	// Introspect with issued Bearer token
	reqIntro := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	reqIntro.Header.Set("Authorization", "Bearer "+verifyResp.Token)
	recIntro := httptest.NewRecorder()
	r.Mux.ServeHTTP(recIntro, reqIntro)

	assert.Equal(t, http.StatusOK, recIntro.Code)
	var claims iam.PrincipalClaims
	err = json.Unmarshal(recIntro.Body.Bytes(), &claims)
	require.NoError(t, err)

	assert.Equal(t, verifyResp.Identity.ID, claims.IdentityID)
	assert.Equal(t, "admin@example.com", claims.Email)
	assert.Equal(t, "01a02086-04a2-75a7-ba24-12d5872b8c49", claims.ActiveOrganizationID)
	assert.Equal(t, "Acme Corp", claims.OrganizationName)
	assert.Equal(t, "owner", claims.OrganizationRole)

	// Introspect with invalid token -> 401
	reqBadIntro := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	reqBadIntro.Header.Set("Authorization", "Bearer tok_bogusinvalidtoken1234567890")
	recBadIntro := httptest.NewRecorder()
	r.Mux.ServeHTTP(recBadIntro, reqBadIntro)
	assert.Equal(t, http.StatusUnauthorized, recBadIntro.Code)
}

func TestVerifyOTP_Success_InvitedUser_NoMemberships(t *testing.T) {
	r, mail, _ := setupTestApp(t)

	// Clean verifications
	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", "01a02086-04a2-75a7-ba24-1616b586c404").Exec(t.Context())
	require.NoError(t, err)

	// Send code to invited user
	bodySend := []byte(`{"email":"invited@example.com"}`)
	reqSend := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/send", bytes.NewReader(bodySend))
	reqSend.Header.Set("Content-Type", "application/json")
	recSend := httptest.NewRecorder()
	r.Mux.ServeHTTP(recSend, reqSend)
	require.Equal(t, http.StatusOK, recSend.Code)

	validCode := mail.LastCode()
	require.NotEmpty(t, validCode)

	// Verify code
	verifyBody, _ := json.Marshal(iam.VerifyOTPRequest{
		Email: "invited@example.com",
		Code:  validCode,
	})
	reqVerify := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(verifyBody))
	reqVerify.Header.Set("Content-Type", "application/json")
	recVerify := httptest.NewRecorder()
	r.Mux.ServeHTTP(recVerify, reqVerify)

	assert.Equal(t, http.StatusOK, recVerify.Code)
	var verifyResp iam.VerifyOTPResponse
	err = json.Unmarshal(recVerify.Body.Bytes(), &verifyResp)
	require.NoError(t, err)

	assert.NotEmpty(t, verifyResp.Token)
	assert.Equal(t, "01a02086-04a2-75a7-ba24-1616b586c404", verifyResp.Identity.ID)
	assert.Nil(t, verifyResp.ActiveOrganization)
	assert.Len(t, verifyResp.PendingInvitations, 1)
	assert.Equal(t, "inv_test1", verifyResp.PendingInvitations[0].ID)
	assert.Equal(t, "member", verifyResp.PendingInvitations[0].Role)
}
