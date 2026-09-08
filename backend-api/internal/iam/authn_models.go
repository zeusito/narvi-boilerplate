package iam

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// VerifyOTPRequest is the request body for verifying the OTP code.
// We expect IP address, and user agent to be provided since the expectation is for this endpoint to be called by the BFF
type VerifyOTPRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Code      string `json:"code" validate:"required,len=6"`
	IPAddress string `json:"ipAddress" validate:"required,ip"`
	UserAgent string `json:"userAgent" validate:"required"`
}

type SignInResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}
