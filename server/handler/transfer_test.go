package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/artalkjs/artalk/v2/server/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fiberTester interface {
	Test(*http.Request, ...int) (*http.Response, error)
}

func TestTransferImportCleansUploadedFiles(t *testing.T) {
	app, fiberApp := NewApiTestApp()
	defer app.Cleanup()
	handler.Transfer(app.App, fiberApp)

	var admin entity.User
	require.NoError(t, app.Dao().DB().Where("is_admin = ?", true).First(&admin).Error)
	token, err := common.LoginGetUserToken(admin, app.Conf().AppKey, 3600)
	require.NoError(t, err)

	t.Run("successful import", func(t *testing.T) {
		filename := uploadTransferFile(t, fiberApp, token, "[{\"content\":\"transfer success\",\"page_key\":\"/transfer-success\",\"site_name\":\"transfer-success\"}]")
		resp := importTransferFile(t, fiberApp, token, filename)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertFileRemoved(t, filename)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		filename := uploadTransferFile(t, fiberApp, token, "not JSON")
		resp := importTransferFile(t, fiberApp, token, filename)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assertFileRemoved(t, filename)
	})

	t.Run("database rollback", func(t *testing.T) {
		trigger := "\n" +
			"CREATE TRIGGER fail_transfer_import\n" +
			"BEFORE INSERT ON atk_comments\n" +
			"WHEN NEW.content = 'transfer rollback'\n" +
			"BEGIN\n" +
			"  SELECT RAISE(ABORT, 'injected transfer failure');\n" +
			"END;\n"
		require.NoError(t, app.Dao().DB().Exec(trigger).Error)
		defer app.Dao().DB().Exec("DROP TRIGGER fail_transfer_import")

		filename := uploadTransferFile(t, fiberApp, token, "[{\"content\":\"transfer rollback\",\"page_key\":\"/transfer-rollback\",\"site_name\":\"transfer-rollback\"}]")
		resp := importTransferFile(t, fiberApp, token, filename)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assertFileRemoved(t, filename)

		var count int64
		require.NoError(t, app.Dao().DB().Model(&entity.Site{}).Where("name = ?", "transfer-rollback").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("file cannot be reused", func(t *testing.T) {
		filename := uploadTransferFile(t, fiberApp, token, "[{\"content\":\"transfer once\",\"page_key\":\"/transfer-once\",\"site_name\":\"transfer-once\"}]")
		resp := importTransferFile(t, fiberApp, token, filename)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = importTransferFile(t, fiberApp, token, filename)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func uploadTransferFile(t *testing.T, fiberApp fiberTester, token, content string) string {
	t.Helper()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "backup.artrans")
	require.NoError(t, err)
	_, err = io.Copy(part, strings.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, w.WriteField("token", token))
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/transfer/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var payload handler.ResponseTransferUpload
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload.Filename)
	require.FileExists(t, payload.Filename)
	return payload.Filename
}

func importTransferFile(t *testing.T, fiberApp fiberTester, token, filename string) *http.Response {
	t.Helper()

	form := url.Values{
		"json_file": {filename},
		"token":     {token},
	}
	req := httptest.NewRequest(http.MethodPost, "/transfer/import", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)
	return resp
}

func assertFileRemoved(t *testing.T, filename string) {
	t.Helper()
	_, err := os.Stat(filename)
	assert.True(t, os.IsNotExist(err), fmt.Sprintf("temporary file %q should be removed", filename))
}
