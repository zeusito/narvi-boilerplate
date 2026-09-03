package iam

import "time"

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SendOTPResponse struct {
	Sent bool `json:"sent"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type IdentityDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	State     string `json:"state"`
}

type ActiveOrganizationDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Logo string `json:"logo"`
	Role string `json:"role"`
}

type PendingInvitationDTO struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	TargetID  string    `json:"target_id"`
	Role      string    `json:"role"`
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expires_at"`
}

type VerifyOTPResponse struct {
	Token              string                 `json:"token"`
	Identity           IdentityDTO            `json:"identity"`
	ActiveOrganization *ActiveOrganizationDTO `json:"active_organization"`
	PendingInvitations []PendingInvitationDTO `json:"pending_invitations"`
}

type PrincipalClaims struct {
	SessionID            string `json:"session_id"`
	IdentityID           string `json:"identity_id"`
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	ActiveOrganizationID string `json:"active_organization_id,omitempty"`
	OrganizationName     string `json:"organization_name,omitempty"`
	OrganizationSlug     string `json:"organization_slug,omitempty"`
	OrganizationLogo     string `json:"organization_logo,omitempty"`
	OrganizationRole     string `json:"organization_role,omitempty"`
}
