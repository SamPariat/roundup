package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLoadWithoutFileIsNoOp(t *testing.T) {
	// No .env.beta file exists in this package's directory, so loading must not fail.
	t.Setenv(AppEnvKey, Beta)
	assert.Equal(t, Beta, Load())
}

func TestAppEnv(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		unset   bool
		want    string
		wantErr bool
	}{
		{name: "Beta", value: "beta", want: Beta},
		{name: "Prod", value: "prod", want: Prod},
		{name: "UnsetIsAnError", unset: true, wantErr: true},
		{name: "EmptyIsAnError", value: "", wantErr: true},
		{name: "LongFormRejected", value: "production", wantErr: true},
		{name: "CaseSensitive", value: "Prod", wantErr: true},
		{name: "NoTrimming", value: " beta", wantErr: true},
		{name: "UnknownEnvironment", value: "staging", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(AppEnvKey, tt.value) // restores the original value after the test
			if tt.unset {
				require.NoError(t, os.Unsetenv(AppEnvKey))
			}

			got, err := appEnv()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), AppEnvKey, "error must name the variable")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadPanicsWithoutAppEnv(t *testing.T) {
	t.Setenv(AppEnvKey, "")
	require.NoError(t, os.Unsetenv(AppEnvKey))

	assert.Panics(t, func() { Load() })
}
