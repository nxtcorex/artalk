package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/artalkjs/artalk/v2/server/handler"
	"github.com/stretchr/testify/require"
)

func TestOverview(t *testing.T) {
	app, router := NewApiTestApp()
	defer app.Cleanup()
	handler.Overview(app.App, router)
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
		{"", "", 403, 0},
		{"", userToken, 403, 0},
		{"", adminToken, 200, 7},
		{"?site_name=Site%20A&days=30", adminToken, 200, 30},
		{"?days=365", adminToken, 400, 0},
		{"?days=oops", adminToken, 400, 0},
		{"?site_name=missing", adminToken, 404, 0},
	} {
		t.Run(tc.query+tc.token[:min(len(tc.token), 6)], func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin/overview"+tc.query, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp, err := router.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tc.code, resp.StatusCode)
			if tc.code == 200 {
				var result dao.Overview
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
				require.Len(t, result.Trend, tc.days)
				require.NotEmpty(t, result.GeneratedAt)
			}
		})
	}
}
