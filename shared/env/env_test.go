package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		envValue   string
		setEnv     bool
		defaultVal string
		want       string
	}{
		{name: "ReturnsEnvValueWhenSet", key: "ROUNDUP_TEST_GET", envValue: "from-env", setEnv: true, defaultVal: "fallback", want: "from-env"},
		{name: "ReturnsDefaultWhenUnset", key: "ROUNDUP_TEST_GET_UNSET", setEnv: false, defaultVal: "fallback", want: "fallback"},
		{name: "ReturnsDefaultWhenSetButEmpty", key: "ROUNDUP_TEST_GET_EMPTY", envValue: "", setEnv: true, defaultVal: "fallback", want: "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			assert.Equal(t, tt.want, Get(tt.key, tt.defaultVal))
		})
	}
}

func TestRequire(t *testing.T) {
	t.Run("ReturnsValueWhenSet", func(t *testing.T) {
		t.Setenv("ROUNDUP_TEST_REQUIRE", "value")
		assert.Equal(t, "value", Require("ROUNDUP_TEST_REQUIRE"))
	})

	t.Run("PanicsWhenUnset", func(t *testing.T) {
		assert.Panics(t, func() { Require("ROUNDUP_TEST_REQUIRE_UNSET") })
	})

	t.Run("PanicsWhenSetButEmpty", func(t *testing.T) {
		t.Setenv("ROUNDUP_TEST_REQUIRE_EMPTY", "")
		assert.Panics(t, func() { Require("ROUNDUP_TEST_REQUIRE_EMPTY") })
	})
}

func TestLoad(t *testing.T) {
	// No .env file exists in this package's directory, so Load must be a silent no-op.
	assert.NotPanics(t, Load)
}
