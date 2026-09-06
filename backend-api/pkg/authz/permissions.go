package authz

import "strings"

// System Roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// System Permissions
const (
	// Organization permissions
	PermOrgRead   = "org:read"
	PermOrgUpdate = "org:update"
	PermOrgDelete = "org:delete"

	// Member permissions
	PermMembersRead       = "members:read"
	PermMembersInvite     = "members:invite"
	PermMembersUpdateRole = "members:update_role"
	PermMembersRemove     = "members:remove"

	// Invitation permissions
	PermInvitationsRead   = "invitations:read"
	PermInvitationsRevoke = "invitations:revoke"

	// Domain resource permissions
	PermResourcesRead   = "resources:read"
	PermResourcesCreate = "resources:create"
	PermResourcesUpdate = "resources:update"
	PermResourcesDelete = "resources:delete"
)

// rolePermissionsMap maps canonical system roles to their granted permissions.
var rolePermissionsMap = map[string][]string{
	RoleOwner: {
		PermOrgRead, PermOrgUpdate, PermOrgDelete,
		PermMembersRead, PermMembersInvite, PermMembersUpdateRole, PermMembersRemove,
		PermInvitationsRead, PermInvitationsRevoke,
		PermResourcesRead, PermResourcesCreate, PermResourcesUpdate, PermResourcesDelete,
	},
	RoleAdmin: {
		PermOrgRead, PermOrgUpdate,
		PermMembersRead, PermMembersInvite, PermMembersUpdateRole, PermMembersRemove,
		PermInvitationsRead, PermInvitationsRevoke,
		PermResourcesRead, PermResourcesCreate, PermResourcesUpdate, PermResourcesDelete,
	},
	RoleMember: {
		PermOrgRead,
		PermMembersRead,
		PermResourcesRead, PermResourcesCreate, PermResourcesUpdate,
	},
	RoleViewer: {
		PermOrgRead,
		PermMembersRead,
		PermResourcesRead,
	},
}

// ExpandRolePermissions expands a role string into its corresponding list of permissions.
// The role check is case-insensitive. If the role is unrecognized or empty, an empty slice is returned.
func ExpandRolePermissions(role string) []string {
	normalized := strings.ToLower(strings.TrimSpace(role))
	perms, ok := rolePermissionsMap[normalized]
	if !ok {
		return []string{}
	}

	result := make([]string, len(perms))
	copy(result, perms)
	return result
}
