package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/terrors"
	"backend-api/pkg/toolbox"
	"backend-api/pkg/toolbox/hasher"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type authnUseCases struct {
	identityRepo     identityRepository
	verificationRepo verificationRepository
	sessionRepo      sessionRepository
	mailer           mailer.Mailer
	hmacHasher       hasher.Hasher
}

func newAuthnUseCases(
	identityRepo identityRepository,
	verificationRepo verificationRepository,
	sessionRepo sessionRepository,
	mail mailer.Mailer,
	hmacHash hasher.Hasher,
) authUseCases {
	return &authnUseCases{
		identityRepo:     identityRepo,
		verificationRepo: verificationRepo,
		sessionRepo:      sessionRepo,
		mailer:           mail,
		hmacHasher:       hmacHash,
	}
}

func (u *authnUseCases) SendOTP(ctx context.Context, req *SendOTPRequest) (*SendOTPResponse, error) {
	identity, err := u.identityRepo.FindActiveByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// Anti-enumeration: if identity doesn't exist or is not active, return sent=true silently
	if identity == nil {
		return &SendOTPResponse{Sent: true}, nil
	}

	// Verify identity has at least 1 membership or pending invitation
	memberships, err := u.identityRepo.FindMembershipsByIdentityID(ctx, identity.ID)
	if err != nil {
		return nil, err
	}
	invitations, err := u.identityRepo.FindPendingInvitationsByEmail(ctx, identity.Email)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 && len(invitations) == 0 {
		return &SendOTPResponse{Sent: true}, nil
	}

	// Check 60-second cooldown on recent active verification
	activeVerification, err := u.verificationRepo.FindLatestActive(ctx, identity.ID, VerificationKindEmailOTP)
	if err != nil {
		return nil, err
	}
	if activeVerification != nil && time.Since(activeVerification.CreatedAt) < 60*time.Second {
		return nil, terrors.TooManyRequests("please wait before requesting another verification code")
	}

	// Invalidate any prior pending OTPs for this identity
	if err := u.verificationRepo.DeleteAllForIdentityAndKind(ctx, identity.ID, VerificationKindEmailOTP); err != nil {
		return nil, err
	}

	// Generate 6-digit OTP
	code, err := generateOTPCode()
	if err != nil {
		return nil, terrors.OperationFailed("failed to generate verification code")
	}

	// Hash OTP code using HMAC-SHA256
	hashData := fmt.Sprintf("%s:%s", identity.ID, code)
	hashedID, err := u.hmacHasher.Hash(hashData)
	if err != nil {
		return nil, terrors.OperationFailed("failed to hash verification code")
	}

	verification := &Verification{
		ID:         hashedID,
		IdentityID: identity.ID,
		Kind:       VerificationKindEmailOTP,
		Attempts:   0,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		CreatedAt:  time.Now(),
	}

	if err := u.verificationRepo.Create(ctx, verification); err != nil {
		return nil, err
	}

	if err := u.mailer.SendOTPCode(ctx, identity.Email, code); err != nil {
		return nil, terrors.OperationFailed("failed to dispatch verification email")
	}

	return &SendOTPResponse{Sent: true}, nil
}

func (u *authnUseCases) VerifyOTP(ctx context.Context, req *VerifyOTPRequest, ipAddress, userAgent string) (*VerifyOTPResponse, error) {
	identity, err := u.identityRepo.FindActiveByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return nil, terrors.UnAuthorized("invalid credentials")
	}

	v, err := u.verificationRepo.FindLatestActive(ctx, identity.ID, VerificationKindEmailOTP)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, terrors.UnAuthorized("invalid or expired verification code")
	}

	if v.Attempts >= 3 {
		_ = u.verificationRepo.Delete(ctx, v.ID)
		return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
	}

	hashData := fmt.Sprintf("%s:%s", identity.ID, req.Code)
	if !u.hmacHasher.Verify(hashData, v.ID) {
		_ = u.verificationRepo.IncrementAttempts(ctx, v.ID)
		if v.Attempts+1 >= 3 {
			_ = u.verificationRepo.Delete(ctx, v.ID)
			return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
		}
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	// Code is valid: delete verification record
	_ = u.verificationRepo.Delete(ctx, v.ID)

	// Mark email verified
	now := time.Now()
	_ = u.identityRepo.UpdateEmailVerifiedAt(ctx, identity.ID, now)

	// Resolve default organization membership (earliest created)
	memberships, err := u.identityRepo.FindMembershipsByIdentityID(ctx, identity.ID)
	if err != nil {
		return nil, err
	}

	var activeOrgDTO *ActiveOrganizationDTO
	var activeOrgID *string
	if len(memberships) > 0 {
		primaryOrgID := memberships[0].OrganizationID
		org, err := u.identityRepo.FindOrganizationByID(ctx, primaryOrgID)
		if err != nil {
			return nil, err
		}
		if org != nil {
			org.Role = memberships[0].Role
			activeOrgDTO = org
			activeOrgID = &primaryOrgID
		}
	}

	// Create session with opaque token
	token, hashedToken, err := toolbox.GenerateOpaqueToken(u.hmacHasher, "tok")
	if err != nil {
		return nil, terrors.OperationFailed("failed to generate session token")
	}

	session := &IdentitySession{
		ID:             hashedToken,
		IdentityID:     identity.ID,
		OrganizationID: activeOrgID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	pendingInvs, err := u.identityRepo.FindPendingInvitationsByEmail(ctx, identity.Email)
	if err != nil {
		return nil, err
	}

	return &VerifyOTPResponse{
		Token: token,
		Identity: IdentityDTO{
			ID:        identity.ID,
			Email:     identity.Email,
			FirstName: identity.FirstName,
			LastName:  identity.LastName,
			State:     identity.State,
		},
		ActiveOrganization: activeOrgDTO,
		PendingInvitations: pendingInvs,
	}, nil
}

func (u *authnUseCases) Introspect(ctx context.Context, token string) (*PrincipalClaims, error) {
	hashedToken, err := u.hmacHasher.Hash(token)
	if err != nil {
		return nil, terrors.UnAuthorized("invalid token")
	}

	view, err := u.sessionRepo.FindActiveSessionIntrospection(ctx, hashedToken)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, terrors.UnAuthorized("session is invalid or expired")
	}

	return &PrincipalClaims{
		SessionID:            view.SessionID,
		IdentityID:           view.IdentityID,
		Email:                view.IdentityEmail,
		FirstName:            view.IdentityFirstName,
		LastName:             view.IdentityLastName,
		ActiveOrganizationID: view.OrganizationID,
		OrganizationName:     view.OrganizationName,
		OrganizationSlug:     view.OrganizationSlug,
		OrganizationLogo:     view.OrganizationLogo,
		OrganizationRole:     view.OrganizationRole,
	}, nil
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
