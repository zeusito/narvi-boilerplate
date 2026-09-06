package authz

default allow := false

# Tenant Match: Principal's active organization must match the target resource's organization_id
tenant_match if {
	input.principal.active_organization_id != ""
	input.principal.active_organization_id == input.resource.organization_id
}

# Permission Check: The requested action must be present in the principal's permissions
has_permission if {
	input.action in input.principal.permissions
}

# Allow when both tenant boundary and permission match
allow if {
	tenant_match
	has_permission
}
