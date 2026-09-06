package iam

import (
	"net/http"
	"time"

	"backend-api/pkg/authz"
	"backend-api/pkg/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
)

type authnController struct {
	authService authService
	sessionSvc  sessionService
}

func newAuthnController(mux *chi.Mux, sessionIntrospector authz.SessionIntrospector, sSvc sessionService, authService authService) *authnController {
	c := &authnController{
		authService: authService,
		sessionSvc:  sSvc,
	}

	// Public routes
	mux.Group(func(r chi.Router) {
		r.With(httprate.LimitBy(10, time.Minute, c.clientIPKey)).Post("/v1/auth/otp/send", c.handleSendOTP)
		r.With(httprate.LimitBy(10, time.Minute, c.clientIPKey)).Post("/v1/auth/otp/verify", c.handleVerifyOTP)
		r.With(authz.RequireAuthMiddleware(sessionIntrospector)).Get("/v1/auth/instrospect", c.handleIntrospect)
		r.With(authz.RequireAuthMiddleware(sessionIntrospector)).Delete("/v1/auth/logout", c.handleLogout)
	})

	return c
}

// clientIPKey is the rate-limit key. middleware.GetClientIP reads the IP
// resolved in step 1; httprate.CanonicalizeIP buckets IPv6 clients by /64.
func (c *authnController) clientIPKey(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}

func (c *authnController) handleSendOTP(w http.ResponseWriter, r *http.Request) {
	body := new(SendOTPRequest)
	if err := router.BindBody(r, body); err != nil {
		log.Error().Err(err).Msg("Failed to bind request body")
		router.RenderError(r.Context(), w, err)
		return
	}

	// Ignore the error here, as we don't want to reveal whether the email exists or not for security reasons.
	_ = c.authService.SendOTP(r.Context(), body)

	router.RenderJSON(r.Context(), w, http.StatusOK, map[string]string{"message": "OTP sent successfully"})
}

func (c *authnController) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	body := new(VerifyOTPRequest)
	if err := router.BindBody(r, body); err != nil {
		log.Error().Err(err).Msg("Failed to bind request body")
		router.RenderError(r.Context(), w, err)
		return
	}

	resp, err := c.authService.VerifyOTP(r.Context(), body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify OTP")
		router.RenderError(r.Context(), w, err)
		return
	}

	router.RenderJSON(r.Context(), w, http.StatusOK, resp)
}

func (c *authnController) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	claims := authz.ExtractClaimsFromContext(r.Context())

	router.RenderJSON(r.Context(), w, http.StatusOK, claims)
}

func (c *authnController) handleLogout(w http.ResponseWriter, r *http.Request) {
	claims := authz.ExtractClaimsFromContext(r.Context())

	_ = c.sessionSvc.Logout(r.Context(), claims.SessionID)

	router.RenderJSON(r.Context(), w, http.StatusOK, router.DefaultSuccessResponseBody())
}
