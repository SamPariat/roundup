// Package authz is the permission check every service shares: Casbin with one
// embedded model, and the rules supplied by the service as a policy CSV. A
// service embeds its own policy.csv and builds an Enforcer in main, typed with
// its own role and action types, so the rules live in a file, never in code.
package authz

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
)

//go:embed model.conf
var modelConf string

// Enforcer answers whether a role may take an action. R and A are the
// service's own role and action types, so it satisfies the service's
// Authorizer port as is.
type Enforcer[R, A ~string] struct {
	enforcer *casbin.Enforcer
}

// New loads the shared model and the given policy, lines of
// "p, <role>, <action>". It fails only if the policy is empty or malformed,
// which is a build-time mistake.
func New[R, A ~string](policyCSV string) (*Enforcer[R, A], error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("load authz model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, stringadapter.NewAdapter(policyCSV))
	if err != nil {
		return nil, fmt.Errorf("load authz policy: %w", err)
	}

	return &Enforcer[R, A]{enforcer: enforcer}, nil
}

// Can reports whether the policy lets role take action. An evaluation error
// counts as a deny.
func (e *Enforcer[R, A]) Can(role R, action A) bool {
	allowed, err := e.enforcer.Enforce(string(role), string(action))

	return err == nil && allowed
}
