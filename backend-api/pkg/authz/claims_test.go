package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractClaimsFromContext_PointerRoundTrip(t *testing.T) {
	original := &PrincipalClaims{
		IsAuthenticated:      true,
		SessionID:            "ses_123",
		IdentityID:           "id_123",
		Email:                "admin@example.com",
		FullName:             "Super Admin",
		ActiveOrganizationID: "org_123",
		OrganizationKind:     "management",
		OrganizationRole:     "owner",
	}

	ctx := AddClaimsToContext(context.Background(), original)
	claims := ExtractClaimsFromContext(ctx)

	assert.True(t, claims.IsAuthenticated)
	assert.Equal(t, "ses_123", claims.SessionID)
	assert.Equal(t, "id_123", claims.IdentityID)
	assert.Equal(t, "admin@example.com", claims.Email)
	assert.Equal(t, "org_123", claims.ActiveOrganizationID)
	assert.Equal(t, "management", claims.OrganizationKind)
	assert.Equal(t, "owner", claims.OrganizationRole)
}

func TestExtractClaimsFromContext_Missing(t *testing.T) {
	claims := ExtractClaimsFromContext(context.Background())

	assert.False(t, claims.IsAuthenticated)
	assert.Empty(t, claims.SessionID)
}

func TestExtractClaimsFromContext_NilPointer(t *testing.T) {
	ctx := AddClaimsToContext(context.Background(), nil)
	claims := ExtractClaimsFromContext(ctx)

	assert.False(t, claims.IsAuthenticated)
}
