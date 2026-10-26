package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/artalkjs/artalk/v2/server/handler"
	"github.com/stretchr/testify/require"
)

func TestCommentActivity(t *testing.T) {
	app, router := NewApiTestApp()
	defer app.Cleanup()
	handler.CommentActivity(app.App, router)
	adminToken, err := common.LoginGetUserToken(app.Dao().FindUserByID(1000), app.Conf().AppKey, 60)
	require.NoError(t, err)
	userToken, err := common.LoginGetUserToken(app.Dao().FindUserByID(1001), app.Conf().AppKey, 60)
	require.NoError(t, err)
	for _, tc := range []struct {
		query string
		token string
		code  int
		days  int
	}{
		{"", "", 403, 0}, {"", userToken, 403, 0}, {"", adminToken, 200, 365},
		{"?site_name=Site%20A&year=2024", adminToken, 200, 366}, {"?year=2025", adminToken, 200, 365},
		{"?year=oops", adminToken, 400, 0}, {"?year=0", adminToken, 400, 0}, {"?year=1969", adminToken, 400, 0},
		{fmt.Sprintf("?year=%d", time.Now().In(app.Dao().Location()).Year()+1), adminToken, 400, 0},
		{"?site_name=missing", adminToken, 404, 0},
	} {
		t.Run(fmt.Sprintf("%s/%d", tc.query, tc.code), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin/overview/activity"+tc.query, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp, err := router.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tc.code, resp.StatusCode)
			if tc.code == 200 {
				var result dao.CommentActivity
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
				require.Len(t, result.Days, tc.days)
				require.NotEmpty(t, result.Years)
			}
		})
	}
}
