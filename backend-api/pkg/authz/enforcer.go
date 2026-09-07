package authz

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/open-policy-agent/opa/v1/rego"
)

//go:embed policy/authz.rego
var DefaultAuthzPolicy string

// PrincipalInput represents the caller's context passed to OPA.
type PrincipalInput struct {
	IsAuthenticated        bool   `json:"isAuthenticated"`
	ActiveOrganizationID   string `json:"activeOrganizationId"`
	ActiveOrganizationKind string `json:"activeOrganizationKind"`
	Role                   string `json:"role"`
}

// ResourceInput represents the target resource context passed to OPA.
type ResourceInput struct {
	Type           string `json:"type"`
	OrganizationID string `json:"organizationId"`
}

// EvalInput encapsulates the full evaluation document provided to OPA.
type EvalInput struct {
	Principal PrincipalInput `json:"principal"`
	Action    string         `json:"action"`
	Resource  ResourceInput  `json:"resource"`
}

// Enforcer defines the policy evaluation interface.
type Enforcer interface {
	IsAllowed(ctx context.Context, input EvalInput) (bool, error)
}

// OPAEnforcer implements Enforcer using an in-process pre-compiled Rego query.
type OPAEnforcer struct {
	query rego.PreparedEvalQuery
}

// NewEnforcer compiles the Rego policy into a prepared evaluation query.
// If policyContent is empty, the default embedded authz policy is used.
func NewEnforcer(ctx context.Context, policyContent string) (*OPAEnforcer, error) {
	if policyContent == "" {
		policyContent = DefaultAuthzPolicy
	}

	query, err := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module("authz.rego", policyContent),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	return &OPAEnforcer{query: query}, nil
}

// IsAllowed evaluates the given input against the compiled Rego policy.
func (e *OPAEnforcer) IsAllowed(ctx context.Context, input EvalInput) (bool, error) {
	results, err := e.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, fmt.Errorf("opa evaluation error: %w", err)
	}

	return results.Allowed(), nil
}
