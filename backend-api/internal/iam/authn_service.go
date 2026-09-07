package iam

import (
	"context"
	"errors"
	"time"

	"backend-api/pkg/mailer"
	"backend-api/pkg/terrors"
	"backend-api/pkg/toolbox"
	"backend-api/pkg/toolbox/hasher"

	"github.com/rs/zerolog/log"
)

type defaultAuthService struct {
	orgRepo          organizationRepository
	identityRepo     identityRepository
	verificationRepo verificationRepository
	sessionRepo      sessionRepository
	mailer           mailer.Mailer
	hmacHasher       hasher.Hasher
}

func newAuthnService(
	orgRepo organizationRepository,
	identityRepo identityRepository,
	verificationRepo verificationRepository,
	sessionRepo sessionRepository,
	mail mailer.Mailer,
	hmacHasher hasher.Hasher,
) authService {
	return &defaultAuthService{
		orgRepo:          orgRepo,
		identityRepo:     identityRepo,
		verificationRepo: verificationRepo,
		sessionRepo:      sessionRepo,
		mailer:           mail,
		hmacHasher:       hmacHasher,
	}
}

// SendOTP sends an OTP to the user's email address. It returns nil for non-existent emails
// and cooldown violations to prevent account enumeration.
func (s *defaultAuthService) SendOTP(ctx context.Context, req *SendOTPRequest) error {
	log.Info().Str("email", req.Email).Msg("sending OTP to email")

	identity, err := s.identityRepo.FindActiveByEmail(ctx, req.Email)
	if err != nil {
		var terr *terrors.Terror
		if errors.As(err, &terr) && terr.ErrCode == "RecordNotFound" {
			log.Info().Str("email", req.Email).Msg("identity not found, ignoring to prevent enumeration")
			return nil
		}
		log.Error().Err(err).Str("email", req.Email).Msg("failed to find active identity by email")
		return terrors.OperationFailed("failed to send verification code")
	}

	// Check 60-second cooldown on recent active verification
	activeVerification, _ := s.verificationRepo.FindLatestActive(ctx, identity.ID, VerificationKindEmailOTP)
	if activeVerification != nil && time.Since(activeVerification.CreatedAt) < 60*time.Second {
		log.Warn().Str("identity_id", identity.ID).Msg("verification request too frequent")
		return nil
	}

	// Invalidate any prior pending OTPs for this identity
	if err := s.verificationRepo.DeleteAllForIdentityAndKind(ctx, identity.ID, VerificationKindEmailOTP); err != nil {
		log.Error().Err(err).Str("identity_id", identity.ID).Msg("failed to delete prior OTPs for identity")
		return err
	}

	// Generate 6-digit OTP
	code, err := toolbox.SecureRandomOTP()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate OTP")
		return terrors.OperationFailed("failed to generate verification code")
	}
	now := time.Now().UTC()

	// Hash OTP code using HMAC-SHA256
	hashedCode, err := s.hmacHasher.Hash(code)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash verification code")
		return terrors.OperationFailed("failed to hash verification code")
	}

	verification := &Verification{
		ID:         toolbox.GenerateTypeId(toolbox.TypeIdPrefixVerification),
		IdentityID: identity.ID,
		Kind:       VerificationKindEmailOTP,
		HashedCode: hashedCode,
		Attempts:   0,
		ExpiresAt:  now.Add(10 * time.Minute),
		CreatedAt:  now,
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		log.Error().Err(err).Str("identity_id", identity.ID).Msg("failed to create verification")
		return err
	}

	if err := s.mailer.SendOTPCode(ctx, identity.Email, code); err != nil {
		log.Error().Err(err).Str("identity_id", identity.ID).Msg("failed to dispatch verification email")
		return terrors.OperationFailed("failed to dispatch verification email")
	}

	return nil
}

// VerifyOTP verifies the OTP code against the stored hash and returns a signed-in response
func (s *defaultAuthService) VerifyOTP(ctx context.Context, req *VerifyOTPRequest) (*SignInResponse, error) {
	now := time.Now().UTC()

	log.Ctx(ctx).Info().Str("email", req.Email).Msg("verifying OTP")

	identity, err := s.identityRepo.FindActiveByEmail(ctx, req.Email)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to find active identity by email")
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	v, err := s.verificationRepo.FindLatestActive(ctx, identity.ID, VerificationKindEmailOTP)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to find latest active verification")
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	if v.Attempts >= 3 {
		log.Ctx(ctx).Warn().Msg("verification attempts exceeded")
		_ = s.verificationRepo.Delete(ctx, v.ID)
		return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
	}

	// Verify the code against the stored hash
	if !s.hmacHasher.Verify(req.Code, v.HashedCode) {
		err = s.verificationRepo.IncrementAttempts(ctx, v.ID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to increment verification attempts")
			return nil, terrors.UnAuthorized("invalid verification code")
		}
		if v.Attempts+1 >= 3 {
			log.Ctx(ctx).Warn().Msg("verification attempts exceeded")
			_ = s.verificationRepo.Delete(ctx, v.ID)
			return nil, terrors.UnAuthorized("verification attempts exceeded, please request a new code")
		}
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	// Code is valid: delete verification record
	err = s.verificationRepo.Delete(ctx, v.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete verification")
		return nil, terrors.UnAuthorized("invalid verification code")
	}

	// Mark email verified
	_ = s.identityRepo.UpdateEmailVerifiedAt(ctx, identity.ID, now)

	// Resolve default organization membership (earliest created)
	activeOrgID := ""
	oldestMembership, _ := s.orgRepo.FindOldestMembershipsByIdentityID(ctx, identity.ID)
	if oldestMembership != nil {
		activeOrgID = oldestMembership.OrganizationID
	}

	// Create session with opaque token
	token, hashedToken, err := toolbox.GenerateOpaqueToken(s.hmacHasher, "ses")
	if err != nil {
		return nil, terrors.OperationFailed("failed to generate session token")
	}

	session := &IdentitySession{
		ID:             hashedToken,
		IdentityID:     identity.ID,
		OrganizationID: activeOrgID,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		ExpiresAt:      now.Add(24 * time.Hour),
		CreatedAt:      now,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &SignInResponse{
		Token:     token,
		ExpiresIn: int(24 * time.Hour / time.Second),
	}, nil
}
