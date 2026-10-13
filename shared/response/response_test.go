package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	got := OK("payload")

	assert.True(t, got.Success)
	assert.Equal(t, "payload", got.Data)
	assert.Empty(t, got.Error)
}

func TestFail(t *testing.T) {
	got := Fail[string]("something went wrong")

	assert.False(t, got.Success)
	assert.Empty(t, got.Data)
	assert.Equal(t, "something went wrong", got.Error)
}
