// Package policy holds event-service's permission rules as data. The rules
// live in policy.csv, not in code; cmd/main.go loads them into the shared
// Casbin enforcer. See docs/PERMISSIONS.md.
package policy

import _ "embed"

// CSV is policy.csv: one "p, <role>, <action>" line per allowed action.
//
//go:embed policy.csv
var CSV string
