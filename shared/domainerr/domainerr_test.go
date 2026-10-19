package domainerr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSentinel = errors.New("thing not found")

func TestNewKeepsSentinelIdentity(t *testing.T) {
	err := New(errSentinel, "id: 42")

	require.ErrorIs(t, err, errSentinel)
	assert.Equal(t, "thing not found: id: 42", err.Error())
}

func TestNewMatchesThroughWrapping(t *testing.T) {
	err := fmt.Errorf("load: %w", New(errSentinel, "id: 42"))

	require.ErrorIs(t, err, errSentinel)

	domainErr, ok := errors.AsType[*Error](err)
	assert.True(t, ok)
	assert.Equal(t, "id: 42", domainErr.Detail)
}
