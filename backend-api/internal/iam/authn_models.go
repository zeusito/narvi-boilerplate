package iam

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

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
