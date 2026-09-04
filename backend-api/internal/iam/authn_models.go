package iam

import "time"

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type SendOTPResponse struct {
	Sent bool `json:"sent"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Code  string `json:"code" validate:"required,len=6"`
}

type IdentityDTO struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	State     IdentityState `json:"state"`
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
