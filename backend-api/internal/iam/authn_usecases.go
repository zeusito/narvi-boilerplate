package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/terrors"
	"backend-api/pkg/toolbox"
	"backend-api/pkg/toolbox/hasher"
	"fmt"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
)

type authnUseCases struct {
	orgRepo          organizationRepository
	identityRepo     identityRepository
	verificationRepo verificationRepository
	sessionRepo      sessionRepository
	mailer           mailer.Mailer
	hmacHasher       hasher.Hasher
}

func newAuthnUseCases(
	orgRepo organizationRepository,
	identityRepo identityRepository,
	verificationRepo verificationRepository,
	sessionRepo sessionRepository,
	mail mailer.Mailer,
	hmacHasher hasher.Hasher,
) authUseCases {
	return &authnUseCases{
		orgRepo:          orgRepo,
		identityRepo:     identityRepo,
		verificationRepo: verificationRepo,
		sessionRepo:      sessionRepo,
		mailer:           mail,
		hmacHasher:       hmacHasher,
	}
}

// SendOTP sends an OTP to the user's email address, it does not return an error to avoid enumeration attacks
func (u *authnUseCases) SendOTP(ctx *echo.Context, req *SendOTPRequest) {
	traceId := toolbox.GetTraceId(ctx)

	log.Info().Str("traceId", traceId).Msgf("sending OTP to email: %s", req.Email)

	identity, err := u.identityRepo.FindActiveByEmail(ctx.Request().Context(), req.Email)
	if err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to find active identity by email")
		return
	}

	// Check 60-second cooldown on recent active verification
	activeVerification, err := u.verificationRepo.FindLatestActive(ctx.Request().Context(), identity.ID, VerificationKindEmailOTP)
	if activeVerification != nil && time.Since(activeVerification.CreatedAt) < 60*time.Second {
		log.Error().Str("traceId", traceId).Msg("verification request too frequent")
		return
	}

	// Invalidate any prior pending OTPs for this identity
	if err := u.verificationRepo.DeleteAllForIdentityAndKind(ctx.Request().Context(), identity.ID, VerificationKindEmailOTP); err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to delete all pending OTPs for identity")
		return
	}

	// Generate 6-digit OTP
	code, err := toolbox.SecureRandomOTP(6)
	if err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to generate OTP")
		return
	}
	now := time.Now().UTC()

	// Hash OTP code using HMAC-SHA256
	hashedID, err := u.hmacHasher.Hash(code)
	if err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to hash verification code")
		return
	}

	verification := &Verification{
		ID:         hashedID,
		IdentityID: identity.ID,
		Kind:       VerificationKindEmailOTP,
		Attempts:   0,
		ExpiresAt:  now.Add(10 * time.Minute),
		CreatedAt:  now,
	}

	if err := u.verificationRepo.Create(ctx.Request().Context(), verification); err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to create verification")
		return
	}

	if err := u.mailer.SendOTPCode(ctx.Request().Context(), identity.Email, code); err != nil {
		log.Error().Str("traceId", traceId).Err(err).Msg("failed to dispatch verification email")
		return
	}
}

func (u *authnUseCases) VerifyOTP(ctx *echo.Context, req *VerifyOTPRequest, ipAddress, userAgent string) (*VerifyOTPResponse, error) {
	traceId := toolbox.GetTraceId(ctx)
	now := time.Now().UTC()

	log.Info().Str("traceId", traceId).Msg("verify otp...")

	identity, err := u.identityRepo.FindActiveByEmail(ctx.Request().Context(), req.Email)
	if err != nil {
		return nil, err
	}

	v, err := u.verificationRepo.FindLatestActive(ctx.Request().Context(), identity.ID, VerificationKindEmailOTP)
	if err != nil {
		return nil, err
	}

	if v.Attempts >= 3 {
		_ = u.verificationRepo.Delete(ctx.Request().Context(), v.ID)
		return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
	}

	// Verify the code against the stored hash
	if !u.hmacHasher.Verify(req.Code, v.ID) {
		_ = u.verificationRepo.IncrementAttempts(ctx.Request().Context(), v.ID)
		if v.Attempts+1 >= 3 {
			_ = u.verificationRepo.Delete(ctx.Request().Context(), v.ID)
			return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
		}
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	// Code is valid: delete verification record
	_ = u.verificationRepo.Delete(ctx.Request().Context(), v.ID)

	// Mark email verified
	_ = u.identityRepo.UpdateEmailVerifiedAt(ctx.Request().Context(), identity.ID, now)

	// Resolve default organization membership (earliest created)
	memberships, err := u.identityRepo.FindMembershipsByIdentityID(ctx.Request().Context(), identity.ID)
	if err != nil {
		return nil, err
	}

	var activeOrgDTO *ActiveOrganizationDTO
	var activeOrgID *string
	if len(memberships) > 0 {
		primaryOrgID := memberships[0].OrganizationID
		org, err := u.identityRepo.FindOrganizationByID(ctx.Request().Context(), primaryOrgID)
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

func (u *authnUseCases) Introspect(ctx *echo.Context, token string) (*PrincipalClaims, error) {
	hashedToken, err := u.hmacHasher.Hash(token)
	if err != nil {
		return nil, terrors.UnAuthorized("invalid token")
	}

	view, err := u.sessionRepo.FindActiveSessionIntrospection(ctx.Request().Context(), hashedToken)
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
		FullName:             fmt.Sprintf("%s %s", view.IdentityFirstName, view.IdentityLastName),
		ActiveOrganizationID: view.OrganizationID,
		OrganizationName:     view.OrganizationName,
		OrganizationSlug:     view.OrganizationSlug,
		OrganizationLogo:     view.OrganizationLogo,
		OrganizationRole:     view.OrganizationRole,
	}, nil
}
