package handler_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/artalkjs/artalk/v2/server/handler"
	"github.com/artalkjs/artalk/v2/server/middleware/limiter"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uploadImage(t *testing.T, router *fiber.App, data []byte, token string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	// The server must choose the extension from the image content.
	part, err := writer.CreateFormFile("file", "untrusted.html")
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := router.Test(req, 5000)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestUploadBackends(t *testing.T) {
	app, router := NewApiTestApp()
	defer app.Cleanup()
	app.Conf().Captcha.Always = true
	router.Use(limiter.ActionLimitMiddleware(app.App, limiter.ActionLimitConf{}))
	handler.Upload(app.App, router)
	adminToken, err := common.LoginGetUserToken(app.Dao().FindUserByID(1000), app.Conf().AppKey, 60)
	require.NoError(t, err)
	var imageData bytes.Buffer
	require.NoError(t, png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1, 1))))
	data := imageData.Bytes()
	var hits, storageStatus atomic.Int32
	storageStatus.Store(http.StatusOK)
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.URL.Path, "/images/comments/")
		assert.Equal(t, "image/png", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, data, body)
		w.WriteHeader(int(storageStatus.Load()))
	}))
	defer storage.Close()

	for _, tc := range []struct {
		name      string
		configure func()
		data      []byte
		token     string
		status    int
		requests  int32
	}{
		{"S3 success", func() {}, data, "", 200, 1},
		{"admin upload when disabled", func() { app.Conf().ImgUpload.Enabled = false }, data, adminToken, 200, 1},
		{"S3 failure", func() { storageStatus.Store(http.StatusForbidden) }, data, "", 500, 1},
		{"unsupported format", func() {}, []byte("<svg>not an allowed image</svg>"), "", 400, 0},
		{"oversized image", func() {}, bytes.Repeat([]byte{0}, 1024*1024+1), "", 400, 0},
		{"upload disabled", func() { app.Conf().ImgUpload.Enabled = false }, data, "", 403, 0},
		{"captcha required", func() { app.Conf().Captcha.Enabled = true }, data, "", 403, 0},
		{"invalid S3 config", func() { app.Conf().ImgUpload.S3.Bucket = "" }, data, "", 500, 0},
		{"conflicting backends", func() { app.Conf().ImgUpload.Upgit.Enabled = true }, data, "", 500, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			localPath := filepath.Join(t.TempDir(), "images")
			app.Conf().ImgUpload = config.ImgUploadConf{
				Enabled: true, Path: localPath, MaxSize: 1,
				S3: config.S3Conf{
					Enabled: true, Endpoint: storage.URL, Bucket: "images", PathStyle: true,
					AccessKeyID: "test-key", SecretAccessKey: "test-secret",
					Prefix: "comments", PublicPath: "https://cdn.example.com",
				},
			}
			app.Conf().Captcha.Enabled = false
			hits.Store(0)
			storageStatus.Store(http.StatusOK)
			tc.configure()
			resp := uploadImage(t, router, tc.data, tc.token)
			require.Equal(t, tc.status, resp.StatusCode)
			assert.Equal(t, tc.requests, hits.Load())
			_, err := os.Stat(localPath)
			assert.True(t, os.IsNotExist(err), "S3 upload must not create the local image directory")
			if tc.status == 200 {
				var result handler.ResponseUpload
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
				assert.Equal(t, "image", result.FileType)
				assert.Regexp(t, `^\d{8}-\d{6}\.\d{3}-[A-Za-z0-9]{16}\.png$`, result.FileName)
				assert.Equal(t, "https://cdn.example.com/comments/"+result.FileName, result.PublicURL)
			} else {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.NotContains(t, string(body), "test-secret")
				assert.NotContains(t, string(body), "test-key")
				assert.NotContains(t, string(body), "public_url")
			}
		})
	}

	t.Run("local storage", func(t *testing.T) {
		app.Conf().Captcha.Enabled = false
		app.Conf().ImgUpload = config.ImgUploadConf{
			Enabled: true, Path: t.TempDir(), PublicPath: "/static/images/",
		}
		hits.Store(0)
		resp := uploadImage(t, router, data, "")
		require.Equal(t, 200, resp.StatusCode)
		var result handler.ResponseUpload
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, "/static/images/"+result.FileName, result.PublicURL)
		stored, err := os.ReadFile(filepath.Join(app.Conf().ImgUpload.Path, result.FileName))
		require.NoError(t, err)
		assert.Equal(t, data, stored)
		assert.Zero(t, hits.Load())
	})
}
