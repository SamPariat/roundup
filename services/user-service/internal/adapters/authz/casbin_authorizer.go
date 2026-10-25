// Package authz decides which squad roles may take which actions, using
// Casbin with the policy embedded from model.conf and policy.csv. The rules
// live in those files; see docs/PERMISSIONS.md.
package authz

import (
	_ "embed"
	"fmt"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
)

var (
	//go:embed model.conf
	modelConf string

	//go:embed policy.csv
	policyCSV string
)

// CasbinAuthorizer implements domain.Authorizer with a Casbin enforcer over
// the embedded policy.
type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
}

// NewCasbinAuthorizer loads the embedded model and policy. It fails only if
// the files are malformed, which is a build-time mistake.
func NewCasbinAuthorizer() (*CasbinAuthorizer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("load authz model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, stringadapter.NewAdapter(policyCSV))
	if err != nil {
		return nil, fmt.Errorf("load authz policy: %w", err)
	}

	return &CasbinAuthorizer{enforcer: enforcer}, nil
}

// Can reports whether the policy lets role take action. An evaluation error
// counts as a deny.
func (a *CasbinAuthorizer) Can(role domain.Role, action domain.Action) bool {
	allowed, err := a.enforcer.Enforce(string(role), string(action))

	return err == nil && allowed
}
