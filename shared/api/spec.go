package api

import _ "embed"

// Spec is the OpenAPI 3 document for every route in this repo
// (openapi.yaml). The gateway serves it at /docs, and each service's router
// test checks its routes against it.
//
//go:embed openapi.yaml
var Spec []byte
