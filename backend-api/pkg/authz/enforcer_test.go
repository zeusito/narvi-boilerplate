package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOPAEnforcer(t *testing.T) {
	ctx := context.Background()
	enforcer, err := NewEnforcer(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, enforcer)

	tests := []struct {
		name     string
		input    EvalInput
		expected bool
	}{
		{
			name: "allowed: matching tenant and permission present",
			input: EvalInput{
				Principal: PrincipalInput{
					IsAuthenticated:        true,
					ActiveOrganizationID:   "org_alpha",
					ActiveOrganizationKind: "standard",
					Role:                   "admin",
				},
				Action: "org:read",
				Resource: ResourceInput{
					Type:           "organization",
					OrganizationID: "org_alpha",
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := enforcer.IsAllowed(ctx, tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, allowed)
		})
	}
}

func TestOPAEnforcer_InvalidPolicy(t *testing.T) {
	ctx := context.Background()
	_, err := NewEnforcer(ctx, "invalid rego policy content %%%")
	require.Error(t, err)
}
