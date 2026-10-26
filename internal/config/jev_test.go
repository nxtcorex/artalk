package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJevConfigFromEnvironment(t *testing.T) {
	t.Setenv("ATK_MODERATOR_JEV_ENABLED", "true")
	t.Setenv("ATK_MODERATOR_JEV_API_KEY", "test-key")
	t.Setenv("ATK_MODERATOR_JEV_MODEL", "jev-custom")
	t.Setenv("ATK_MODERATOR_JEV_THRESHOLD", "0.65")
	t.Setenv("ATK_MODERATOR_JEV_HEADERS", `{"CF-Access-Client-Secret":"access-secret"}`)
	file := filepath.Join(t.TempDir(), "artalk.yml")
	require.NoError(t, os.WriteFile(file, []byte("moderator:\n  jev:\n    enabled: false\n    threshold: 0.8\n    headers: {}\n"), 0o600))
	conf, err := NewFromFile(file)
	require.NoError(t, err)
	assert.True(t, conf.Moderator.Jev.Enabled)
	assert.Equal(t, "test-key", conf.Moderator.Jev.APIKey)
	assert.Equal(t, "jev-custom", conf.Moderator.Jev.Model)
	require.NotNil(t, conf.Moderator.Jev.Threshold)
	assert.InDelta(t, 0.65, *conf.Moderator.Jev.Threshold, 0.0001)
	assert.Equal(t, HTTPHeaders{"CF-Access-Client-Secret": "access-secret"}, conf.Moderator.Jev.Headers)

	t.Setenv("ATK_MODERATOR_JEV_THRESHOLD", "0")
	conf, err = NewFromFile(file)
	require.NoError(t, err)
	require.NotNil(t, conf.Moderator.Jev.Threshold)
	assert.Zero(t, *conf.Moderator.Jev.Threshold)
}
