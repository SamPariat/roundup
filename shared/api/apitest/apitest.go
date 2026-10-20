// Package apitest checks a service's router against the OpenAPI spec in
// api.Spec, so a route can't be added or removed without updating the docs.
package apitest

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// httpMethods are the keys of an OpenAPI path item that are operations;
// the others (parameters, servers, ...) are skipped.
var httpMethods = []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

// paramPattern matches a Fiber path parameter such as ":placeID".
var paramPattern = regexp.MustCompile(`:(\w+)`)

// Operations returns every operation in api.Spec as "METHOD /path", e.g.
// "GET /api/v1/venues/{placeID}".
func Operations(t *testing.T) map[string]bool {
	t.Helper()

	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(api.Spec, &doc))

	ops := make(map[string]bool)

	for path, item := range doc.Paths {
		for key := range item {
			if slices.Contains(httpMethods, key) {
				ops[strings.ToUpper(key)+" "+path] = true
			}
		}
	}

	return ops
}

// RouteKey converts a Fiber route to the key Operations uses, e.g.
// ("GET", "/api/v1/venues/:placeID") -> "GET /api/v1/venues/{placeID}".
// Group routes registered as "/" end in a slash ("/api/v1/users/me/"), which
// the spec never does, so it is trimmed.
func RouteKey(method, fiberPath string) string {
	path := paramPattern.ReplaceAllString(fiberPath, "{$1}")
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}

	return method + " " + path
}

// AssertRoutesDocumented fails t unless app's routes under prefix and the
// spec's operations under prefix are the same set. HEAD routes Fiber adds for
// every GET are ignored.
func AssertRoutesDocumented(t *testing.T, app *fiber.App, prefix string) {
	t.Helper()

	routes := make(map[string]bool)

	for _, r := range app.GetRoutes(true) {
		if r.Method == fiber.MethodHead || !strings.HasPrefix(r.Path, prefix) {
			continue
		}

		routes[RouteKey(r.Method, r.Path)] = true
	}

	documented := make(map[string]bool)

	for op := range Operations(t) {
		if _, path, _ := strings.Cut(op, " "); strings.HasPrefix(path, prefix) {
			documented[op] = true
		}
	}

	for route := range routes {
		assert.True(t, documented[route], "route %q is not in shared/api/openapi.yaml", route)
	}

	for op := range documented {
		assert.True(t, routes[op], "openapi.yaml documents %q but the router has no such route", op)
	}
}
