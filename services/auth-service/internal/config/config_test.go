package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequired(t *testing.T, appEnv string) {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)

	t.Setenv("APP_ENV", appEnv)
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("USER_SERVICE_URL", "http://localhost:3002")
	t.Setenv("JWT_PRIVATE_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})))
}

func TestLoadDefaults(t *testing.T) {
	setRequired(t, "beta")

	cfg := Load()

	assert.Equal(t, "beta", cfg.Env)
	assert.Equal(t, "3001", cfg.Port)
	assert.Equal(t, 15*time.Minute, cfg.AccessTokenTTL)
	assert.Equal(t, 720*time.Hour, cfg.RefreshTokenTTL)
	assert.NotNil(t, cfg.JWTPrivateKey)
}

func TestLoadAllowsDevCodeInBeta(t *testing.T) {
	setRequired(t, "beta")
	t.Setenv("OTP_DEV_CODE", "123456")

	assert.Equal(t, "123456", Load().OTPDevCode)
}

func TestLoadPanics(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"DevCodeInProd", map[string]string{"APP_ENV": "prod", "OTP_DEV_CODE": "123456"}},
		{"BadPrivateKey", map[string]string{"JWT_PRIVATE_KEY": "not a key"}},
		{"BadTTL", map[string]string{"ACCESS_TOKEN_TTL": "fifteen minutes"}},
		{"NegativeTTL", map[string]string{"REFRESH_TOKEN_TTL": "-1h"}},
		{"MissingUserServiceURL", map[string]string{"USER_SERVICE_URL": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequired(t, "beta")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			assert.Panics(t, func() { Load() })
		})
	}
}
