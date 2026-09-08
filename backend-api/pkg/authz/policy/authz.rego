package authz

default allow := false

# Check if principal is authenticated
is_authenticated if {
	input.principal.isAuthenticated == true
}

# 1. Management Admin Rule:
# Must be authenticated, active org kind must be "management", and member role must be "admin"
is_management_admin if {
	is_authenticated
	input.principal.activeOrganizationKind == "management"
	input.principal.role == "admin"
}

# Allow management admin on management-scoped resources/actions
allow if {
	is_management_admin
	input.resource.type == "management"
}

# Enforce tenant boundary for normal tenant operations.
# Permissions were removed for now, so any authenticated member of the
# matching tenant is allowed.
tenant_match if {
	input.principal.activeOrganizationId != ""
	input.principal.activeOrganizationId == input.resource.organizationId
}

allow if {
	is_authenticated
	tenant_match
}
