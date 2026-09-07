package authz

import future.keywords.in

default allow := false

# Check if principal is authenticated
is_authenticated if {
	input.principal.is_authenticated == true
}

# 1. Management Admin Rule:
# Must be authenticated, active org kind must be "management", and member role must be "admin"
is_management_admin if {
	is_authenticated
	input.principal.active_organization_kind == "management"
	input.principal.role == "admin"
}

# Allow management admin on management-scoped resources/actions
allow if {
	is_management_admin
	input.resource.type == "management"
}

# Enforce tenant boundary and granular permissions for normal tenant operations
tenant_match if {
	input.principal.active_organization_id != ""
	input.principal.active_organization_id == input.resource.organization_id
}

has_permission if {
	input.action in input.principal.permissions
}

allow if {
	tenant_match
	has_permission
}
