package iam

import (
	"backend-api/pkg/authz"
	"backend-api/pkg/router"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type defaultOrgController struct {
	orgSvc orgService
}

func newOrgController(mux *chi.Mux, sessionIntrospector authz.SessionIntrospector, orgSvc orgService) *defaultOrgController {
	c := &defaultOrgController{orgSvc: orgSvc}

	// Protected routes
	mux.Group(func(r chi.Router) {
		r.Use(authz.RequireAuthMiddleware(sessionIntrospector))
		r.Post("/v1/admin/organizations", c.createOrganization)
		r.Get("/v1/admin/organizations", c.listOrganizations)
		r.Get("/v1/admin/organizations/{id}", c.getOrganizationByID)
	})

	return c
}

func (c *defaultOrgController) createOrganization(w http.ResponseWriter, r *http.Request) {
	body := new(CreateOrganizationRequest)
	if err := router.BindBody(r, body); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to bind request body")
		router.RenderError(r.Context(), w, err)
		return
	}

	resp, err := c.orgSvc.Create(r.Context(), body)
	if err != nil {
		router.RenderError(r.Context(), w, err)
		return
	}

	router.RenderJSON(r.Context(), w, http.StatusCreated, resp)
}

func (c *defaultOrgController) listOrganizations(w http.ResponseWriter, r *http.Request) {
	resp := c.orgSvc.GetAll(r.Context())

	router.RenderJSON(r.Context(), w, http.StatusOK, resp)
}

func (c *defaultOrgController) getOrganizationByID(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")

	resp, err := c.orgSvc.GetById(r.Context(), orgID)
	if err != nil {
		router.RenderError(r.Context(), w, err)
		return
	}

	router.RenderJSON(r.Context(), w, http.StatusOK, resp)
}
