package anti_spam

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMCheckerDecision(t *testing.T) {
	for _, tc := range []struct {
		name     string
		decision string
		pass     bool
	}{
		{name: "allow", decision: "allow", pass: true},
		{name: "review", decision: "review", pass: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v1/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "access-id", r.Header.Get("CF-Access-Client-Id"))
				assert.Equal(t, "access-secret", r.Header.Get("CF-Access-Client-Secret"))

				var request llmChatRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				assert.Equal(t, "test-model", request.Model)
				require.Len(t, request.Messages, 2)
				assert.Contains(t, request.Messages[0].Content, "Custom site policy")
				assert.Contains(t, request.Messages[1].Content, `"comment":"hello"`)
				assert.Equal(t, "json_object", request.ResponseFormat.Type)

				result, err := json.Marshal(llmModerationResponse{
					Decision: tc.decision,
					Reason:   "test reason",
				})
				require.NoError(t, err)
				writeChatCompletion(t, w, string(result))
			}))
			defer server.Close()

			checker := newTestLLMChecker(server.URL + "/v1/chat/completions")
			pass, err := checker.Check(&CheckerParams{Content: "hello"})

			require.NoError(t, err)
			assert.Equal(t, tc.pass, pass)
		})
	}
}

func TestLLMCheckerInvalidResults(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errText string
	}{
		{name: "invalid JSON", content: "not-json", errText: "decode LLM moderation result"},
		{name: "unknown decision", content: `{"decision":"delete"}`, errText: "unknown LLM moderation decision"},
		{name: "unknown field", content: `{"decision":"allow","extra":true}`, errText: "unknown field"},
		{name: "multiple values", content: `{"decision":"allow"}{}`, errText: "multiple JSON values"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeChatCompletion(t, w, tc.content)
			}))
			defer server.Close()

			pass, err := newTestLLMChecker(server.URL).Check(&CheckerParams{Content: "hello"})

			assert.False(t, pass)
			require.ErrorContains(t, err, tc.errText)
		})
	}
}

func TestLLMCheckerRequestFailures(t *testing.T) {
	for _, tc := range []struct {
		name       string
		statusCode int
		message    string
	}{
		{name: "rate limit", statusCode: http.StatusTooManyRequests, message: "rate limited"},
		{name: "server error", statusCode: http.StatusBadGateway, message: "provider unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{"message": tc.message},
				})
			}))
			defer server.Close()

			_, err := newTestLLMChecker(server.URL).Check(&CheckerParams{Content: "hello"})
			require.ErrorContains(t, err, tc.message)
		})
	}

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			writeChatCompletion(t, w, `{"decision":"allow"}`)
		}))
		defer server.Close()

		checker := newTestLLMChecker(server.URL)
		checker.client.Timeout = 10 * time.Millisecond
		_, err := checker.Check(&CheckerParams{Content: "hello"})
		require.ErrorContains(t, err, "request LLM")
	})

	t.Run("oversized response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Repeat("x", maxLLMResponseBytes+1)))
		}))
		defer server.Close()

		_, err := newTestLLMChecker(server.URL).Check(&CheckerParams{Content: "hello"})
		require.ErrorContains(t, err, "response exceeds")
	})
}

func TestLLMCheckerInputLimit(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writeChatCompletion(t, w, `{"decision":"allow"}`)
	}))
	defer server.Close()

	checker := newTestLLMChecker(server.URL)
	checker.conf.MaxInputChars = 3
	_, err := checker.Check(&CheckerParams{Content: "评论过长"})

	require.ErrorContains(t, err, "input limit")
	assert.Zero(t, requests.Load())
}

func TestLLMCheckerConfigValidation(t *testing.T) {
	checker := newTestLLMChecker("ftp://example.com/chat")
	_, err := checker.Check(&CheckerParams{Content: "hello"})
	require.ErrorContains(t, err, "HTTP(S) URL")

	checker = newTestLLMChecker("https://example.com/chat")
	checker.conf.Model = ""
	_, err = checker.Check(&CheckerParams{Content: "hello"})
	require.ErrorContains(t, err, "model is required")
}

func newTestLLMChecker(endpoint string) *LLMChecker {
	return NewLLMChecker(config.LLMAntispamConf{
		Endpoint: endpoint,
		APIKey:   "secret",
		Headers: map[string]string{
			"Authorization":           "Custom auth is overridden by api_key",
			"Content-Type":            "text/plain",
			"CF-Access-Client-Id":     "access-id",
			"CF-Access-Client-Secret": "access-secret",
		},
		Model:  "test-model",
		Policy: "Custom site policy",
	}).(*LLMChecker)
}

func writeChatCompletion(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
		"id":    "chatcmpl-test",
		"model": "test-model",
		"choices": []any{
			map[string]any{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	}))
}
