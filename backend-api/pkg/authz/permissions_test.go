package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandRolePermissions(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		expectedCount  int
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:          "owner has all permissions including org:delete",
			role:          RoleOwner,
			expectedCount: 13,
			mustContain: []string{
				PermOrgRead, PermOrgUpdate, PermOrgDelete,
				PermMembersRead, PermMembersInvite, PermMembersUpdateRole, PermMembersRemove,
				PermInvitationsRead, PermInvitationsRevoke,
				PermResourcesRead, PermResourcesCreate, PermResourcesUpdate, PermResourcesDelete,
			},
			mustNotContain: nil,
		},
		{
			name:          "admin has 12 permissions, missing org:delete",
			role:          RoleAdmin,
			expectedCount: 12,
			mustContain: []string{
				PermOrgRead, PermOrgUpdate,
				PermMembersRead, PermMembersInvite, PermMembersUpdateRole, PermMembersRemove,
				PermInvitationsRead, PermInvitationsRevoke,
				PermResourcesRead, PermResourcesCreate, PermResourcesUpdate, PermResourcesDelete,
			},
			mustNotContain: []string{PermOrgDelete},
		},
		{
			name:          "member has standard CRUD permissions",
			role:          RoleMember,
			expectedCount: 5,
			mustContain: []string{
				PermOrgRead, PermMembersRead, PermResourcesRead, PermResourcesCreate, PermResourcesUpdate,
			},
			mustNotContain: []string{
				PermOrgUpdate, PermOrgDelete, PermMembersInvite, PermMembersRemove, PermMembersUpdateRole, PermInvitationsRevoke, PermResourcesDelete,
			},
		},
		{
			name:          "viewer has only read permissions",
			role:          RoleViewer,
			expectedCount: 3,
			mustContain: []string{
				PermOrgRead, PermMembersRead, PermResourcesRead,
			},
			mustNotContain: []string{
				PermOrgUpdate, PermOrgDelete, PermMembersInvite, PermResourcesCreate, PermResourcesUpdate, PermResourcesDelete,
			},
		},
		{
			name:          "case-insensitivity support",
			role:          "Admin",
			expectedCount: 12,
			mustContain:   []string{PermOrgRead, PermOrgUpdate},
			mustNotContain: []string{PermOrgDelete},
		},
		{
			name:           "unrecognized role returns empty slice",
			role:           "guest",
			expectedCount:  0,
			mustContain:    nil,
			mustNotContain: []string{PermOrgRead},
		},
		{
			name:           "empty role returns empty slice",
			role:           "",
			expectedCount:  0,
			mustContain:    nil,
			mustNotContain: []string{PermOrgRead},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perms := ExpandRolePermissions(tc.role)
			assert.Len(t, perms, tc.expectedCount)

			for _, p := range tc.mustContain {
				assert.Contains(t, perms, p)
			}
			for _, p := range tc.mustNotContain {
				assert.NotContains(t, perms, p)
			}
		})
	}
}
