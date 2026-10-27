package imgupload

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isolateAWSEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "credentials"))
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}

func TestUploadS3(t *testing.T) {
	isolateAWSEnvironment(t)
	for _, tc := range []struct {
		name       string
		prefix     string
		publicPath string
		key        string
		publicURL  string
	}{
		{"bucket root", "", "https://cdn.example.com", "test.png", "https://cdn.example.com/test.png"},
		{"nested prefix", "/images/comments/", "https://cdn.example.com/base/", "images/comments/test.png", "https://cdn.example.com/base/images/comments/test.png"},
		{"escaped prefix", "图片/a b#%", "https://cdn.example.com/base%20path/", "图片/a b#%/test.png", "https://cdn.example.com/base%20path/%E5%9B%BE%E7%89%87/a%20b%23%25/test.png"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte("image bytes")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/images/"+tc.key, r.URL.Path)
				assert.Equal(t, "image/png", r.Header.Get("Content-Type"))
				assert.EqualValues(t, len(data), r.ContentLength)
				assert.Contains(t, r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=test-key/")
				assert.Contains(t, r.Header.Get("Authorization"), "/us-east-1/s3/aws4_request")
				assert.Equal(t, "test-token", r.Header.Get("X-Amz-Security-Token"))
				assert.Empty(t, r.Header.Get("X-Amz-Acl"))
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Equal(t, data, body)
				w.Header().Set("ETag", `"test-etag"`)
			}))
			defer server.Close()
			result, err := UploadS3(context.Background(), config.S3Conf{
				Endpoint: server.URL, Bucket: "images", PathStyle: true,
				AccessKeyID: "test-key", SecretAccessKey: "test-secret", SessionToken: "test-token",
				Prefix: tc.prefix, PublicPath: tc.publicPath,
			}, "test.png", "image/png", data)
			require.NoError(t, err)
			assert.Equal(t, tc.publicURL, result)
		})
	}
}

func TestUploadS3DefaultCredentials(t *testing.T) {
	isolateAWSEnvironment(t)
	t.Setenv("AWS_ACCESS_KEY_ID", "environment-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "environment-secret")
	t.Setenv("AWS_SESSION_TOKEN", "environment-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header.Get("Authorization"), "Credential=environment-key/")
		assert.Contains(t, r.Header.Get("Authorization"), "/eu-west-1/s3/aws4_request")
		assert.Equal(t, "environment-token", r.Header.Get("X-Amz-Security-Token"))
	}))
	defer server.Close()
	result, err := UploadS3(context.Background(), config.S3Conf{
		Endpoint: server.URL, Region: "eu-west-1", Bucket: "images", PathStyle: true,
		PublicPath: "https://cdn.example.com",
	}, "test.jpg", "image/jpeg", []byte("image bytes"))
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/test.jpg", result)
}

func TestUploadS3InvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func(*config.S3Conf)
	}{
		{"missing bucket", func(c *config.S3Conf) { c.Bucket = " " }},
		{"missing public URL", func(c *config.S3Conf) { c.PublicPath = "" }},
		{"relative public URL", func(c *config.S3Conf) { c.PublicPath = "/images/" }},
		{"invalid public URL", func(c *config.S3Conf) { c.PublicPath = "https://%" }},
		{"URL credentials", func(c *config.S3Conf) { c.PublicPath = "https://user:pass@cdn.example.com" }},
		{"URL query", func(c *config.S3Conf) { c.PublicPath += "?token=secret" }},
		{"URL fragment", func(c *config.S3Conf) { c.PublicPath += "#images" }},
		{"endpoint scheme", func(c *config.S3Conf) { c.Endpoint = "file:///tmp/images" }},
		{"endpoint query", func(c *config.S3Conf) { c.Endpoint = "https://s3.example.com?" }},
		{"partial credentials", func(c *config.S3Conf) { c.AccessKeyID = "key" }},
		{"token without credentials", func(c *config.S3Conf) { c.SessionToken = "token" }},
		{"parent path segment", func(c *config.S3Conf) { c.Prefix = "images/../other" }},
		{"dot path segment", func(c *config.S3Conf) { c.Prefix = "images/./other" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := config.S3Conf{Bucket: "images", PublicPath: "https://cdn.example.com"}
			tc.modify(&conf)
			result, err := UploadS3(context.Background(), conf, "test.png", "image/png", []byte("image bytes"))
			require.Error(t, err)
			assert.Empty(t, result)
		})
	}
}

func TestUploadS3Failure(t *testing.T) {
	isolateAWSEnvironment(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "<Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>")
	}))
	defer server.Close()
	result, err := UploadS3(context.Background(), config.S3Conf{
		Endpoint: server.URL, Bucket: "images", PathStyle: true,
		AccessKeyID: "key", SecretAccessKey: "secret", PublicPath: "https://cdn.example.com",
	}, "test.png", "image/png", []byte("image bytes"))
	require.ErrorContains(t, err, "AccessDenied")
	assert.Empty(t, result)
}

func TestUploadS3Timeout(t *testing.T) {
	isolateAWSEnvironment(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	result, err := UploadS3(context.Background(), config.S3Conf{
		Endpoint: server.URL, Bucket: "images", PathStyle: true,
		AccessKeyID: "key", SecretAccessKey: "secret", PublicPath: "https://cdn.example.com",
		Timeout: 1,
	}, "test.png", "image/png", []byte("image bytes"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "error: %v", err)
	assert.Empty(t, result)
}
