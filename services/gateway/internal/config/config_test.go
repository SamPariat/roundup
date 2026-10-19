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

func setRequired(t *testing.T) {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)

	t.Setenv("APP_ENV", "beta")
	t.Setenv("AUTH_SERVICE_URL", "http://localhost:3001")
	t.Setenv("USER_SERVICE_URL", "http://localhost:3002")
	t.Setenv("VENUE_SERVICE_URL", "http://localhost:3006")
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))
}

func TestLoadDefaults(t *testing.T) {
	setRequired(t)

	cfg := Load()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 10*time.Second, cfg.UpstreamTimeout)
	assert.Equal(t, "http://localhost:3002", cfg.Upstreams.User)
	assert.NotNil(t, cfg.JWTPublicKey)
}

func TestLoadPanics(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"MissingVenueURL": {"VENUE_SERVICE_URL": ""},
		"BadPublicKey":    {"JWT_PUBLIC_KEY": "not a key"},
		"BadTimeout":      {"UPSTREAM_TIMEOUT": "ten seconds"},
	} {
		t.Run(name, func(t *testing.T) {
			setRequired(t)

			for k, v := range env {
				t.Setenv(k, v)
			}

			assert.Panics(t, func() { Load() })
		})
	}
}
