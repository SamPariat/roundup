package api

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestSpecIsValidOpenAPI(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData(Spec)
	require.NoError(t, err)

	require.NoError(t, doc.Validate(t.Context()))
}
