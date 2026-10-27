package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Config(t *testing.T) {
	file := filepath.Join(t.TempDir(), "artalk.yml")
	require.NoError(t, os.WriteFile(file, []byte(`img_upload:
  s3:
    enabled: true
    endpoint: https://s3.example.com
    region: eu-west-1
    bucket: images
    access_key_id: file-key
    secret_access_key: file-secret
    session_token: file-token
    prefix: comments
    public_path: https://cdn.example.com
    path_style: true
    timeout: 15
`), 0o600))
	conf, err := NewFromFile(file)
	require.NoError(t, err)
	assert.Equal(t, S3Conf{
		Enabled: true, Endpoint: "https://s3.example.com", Region: "eu-west-1", Bucket: "images",
		AccessKeyID: "file-key", SecretAccessKey: "file-secret", SessionToken: "file-token",
		Prefix: "comments", PublicPath: "https://cdn.example.com", PathStyle: true, Timeout: 15,
	}, conf.ImgUpload.S3)

	for key, value := range map[string]string{
		"ATK_IMG_UPLOAD_S3_ENABLED": "false", "ATK_IMG_UPLOAD_S3_ENDPOINT": "https://storage.example.com",
		"ATK_IMG_UPLOAD_S3_REGION": "us-east-1", "ATK_IMG_UPLOAD_S3_BUCKET": "other-images",
		"ATK_IMG_UPLOAD_S3_ACCESS_KEY_ID": "env-key", "ATK_IMG_UPLOAD_S3_SECRET_ACCESS_KEY": "env-secret",
		"ATK_IMG_UPLOAD_S3_SESSION_TOKEN": "env-token", "ATK_IMG_UPLOAD_S3_PREFIX": "posts",
		"ATK_IMG_UPLOAD_S3_PUBLIC_PATH": "https://img.example.com", "ATK_IMG_UPLOAD_S3_PATH_STYLE": "false",
		"ATK_IMG_UPLOAD_S3_TIMEOUT": "45",
	} {
		t.Setenv(key, value)
	}
	conf, err = NewFromFile(file)
	require.NoError(t, err)
	assert.Equal(t, S3Conf{
		Enabled: false, Endpoint: "https://storage.example.com", Region: "us-east-1", Bucket: "other-images",
		AccessKeyID: "env-key", SecretAccessKey: "env-secret", SessionToken: "env-token",
		Prefix: "posts", PublicPath: "https://img.example.com", PathStyle: false, Timeout: 45,
	}, conf.ImgUpload.S3)
}
