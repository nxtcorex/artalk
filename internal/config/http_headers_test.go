package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPHeadersFromEnvironment(t *testing.T) {
	t.Setenv("ATK_MODERATOR_LLM_HEADERS", `{"CF-Access-Client-Id":"access-id","CF-Access-Client-Secret":"access-secret"}`)

	configFile := filepath.Join(t.TempDir(), "artalk.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("moderator:\n  llm:\n    headers: {}\n"), 0o600))

	conf, err := NewFromFile(configFile)
	require.NoError(t, err)
	assert.Equal(t, HTTPHeaders{
		"CF-Access-Client-Id":     "access-id",
		"CF-Access-Client-Secret": "access-secret",
	}, conf.Moderator.LLM.Headers)
}
