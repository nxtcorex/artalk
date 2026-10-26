package dao_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/internal/db"
	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func overviewDAO(t *testing.T) *dao.Dao {
	t.Helper()
	database, err := db.NewDB(config.DBConf{Type: config.TypeSQLite, File: filepath.Join(t.TempDir(), "overview.db"), TablePrefix: "overview_"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.CloseDB(database)) })
	return dao.NewDao(database)
}

func TestOverviewAggregation(t *testing.T) {
	d := overviewDAO(t)
	location, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	d.SetLocation(location)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, location)
	today := time.Date(2026, 10, 26, 0, 0, 0, 0, location)
	user := entity.User{Name: "Reader", Email: "reader@example.com"}
	admin := entity.User{Name: "Admin", Email: "admin@example.com", IsAdmin: true}
	pendingUser := entity.User{Name: "Pending reader", Email: "pending@example.com"}
	for _, u := range []*entity.User{&user, &admin, &pendingUser} {
		require.NoError(t, d.DB().Create(u).Error)
	}
	for _, site := range []entity.Site{{Name: "A", Urls: "https://a.example"}, {Name: "B", Urls: "https://b.example"}} {
		require.NoError(t, d.DB().Create(&site).Error)
	}
	a := entity.Page{Key: "/shared", Title: "Page A", SiteName: "A", PV: 10}
	b := entity.Page{Key: "/shared", Title: "Page B", SiteName: "B", PV: 20}
	pendingPage := entity.Page{Key: "/pending", SiteName: "A", PV: 3}
	emptyPage := entity.Page{Key: "/empty", SiteName: "A", PV: 7}
	deletedPage := entity.Page{Key: "/deleted", SiteName: "A", PV: 99}
	for _, page := range []*entity.Page{&a, &b, &pendingPage, &emptyPage, &deletedPage} {
		require.NoError(t, d.DB().Create(page).Error)
	}
	require.NoError(t, d.DB().Delete(&deletedPage).Error)
	add := func(page entity.Page, u uint, date time.Time, pending, collapsed bool) entity.Comment {
		c := entity.Comment{Model: gorm.Model{CreatedAt: date.In(time.Local)}, PageKey: page.Key, SiteName: page.SiteName,
			UserID: u, Content: "**Hello** &amp; world<script>alert(1)</script>", IsPending: pending, IsCollapsed: collapsed}
		require.NoError(t, d.DB().Create(&c).Error)
		return c
	}
	add(a, user.ID, today.AddDate(0, 0, -40), false, false)
	add(a, user.ID, today.AddDate(0, 0, -6), false, false)
	add(pendingPage, pendingUser.ID, today.AddDate(0, 0, -1), true, false)
	add(a, user.ID, now, false, false)
	add(a, admin.ID, now, false, true)
	add(a, 999999, now, false, false) // Missing users must not inflate participants.
	latest := add(a, pendingUser.ID, now, true, false)
	deleted := add(emptyPage, user.ID, now, false, false)
	require.NoError(t, d.DB().Delete(&deleted).Error)
	add(b, pendingUser.ID, now, false, false)

	result, err := d.GetOverview("A", 7, now)
	require.NoError(t, err)
	require.Equal(t, dao.OverviewMetrics{TodayComments: 4, Pending: 2, Published: 5, Participants: 2, CommentedPages: 2, PV: 20}, result.Metrics)
	require.Equal(t, "Asia/Shanghai", result.Timezone)
	require.Equal(t, now.Format(time.RFC3339), result.GeneratedAt)
	require.Len(t, result.Trend, 7)
	require.Equal(t, dao.OverviewDay{Date: "2026-10-20", Published: 1}, result.Trend[0])
	require.Equal(t, dao.OverviewDay{Date: "2026-10-21"}, result.Trend[1])
	require.Equal(t, dao.OverviewDay{Date: "2026-10-25", Pending: 1}, result.Trend[5])
	require.Equal(t, dao.OverviewDay{Date: "2026-10-26", Published: 3, Pending: 1}, result.Trend[6])
	require.Len(t, result.LatestComments, 5)
	require.Equal(t, latest.ID, result.LatestComments[0].ID)
	require.True(t, result.LatestComments[0].IsPending)
	require.Equal(t, "https://a.example/shared", result.LatestComments[0].PageURL)
	require.Equal(t, "Page A", result.LatestComments[0].PageTitle)
	require.NotContains(t, result.LatestComments[0].Summary, "<")
	require.NotContains(t, result.LatestComments[0].Summary, "alert")
	require.Contains(t, result.LatestComments[0].Summary, "Hello & world")
	require.Len(t, result.PopularPages, 1)
	require.Equal(t, a.ID, result.PopularPages[0].ID)
	require.EqualValues(t, 4, result.PopularPages[0].Comments)

	all, err := d.GetOverview("", 30, now)
	require.NoError(t, err)
	require.Len(t, all.Trend, 30)
	require.Equal(t, dao.OverviewMetrics{TodayComments: 5, Pending: 2, Published: 6, Participants: 3, CommentedPages: 3, PV: 40}, all.Metrics)
	require.Len(t, all.PopularPages, 2)
	require.Equal(t, "A", all.PopularPages[0].SiteName)
	require.EqualValues(t, 4, all.PopularPages[0].Comments)
	require.Equal(t, "B", all.PopularPages[1].SiteName)
	require.EqualValues(t, 1, all.PopularPages[1].Comments)

	// Current review status, rather than an invented review history, drives trends.
	require.NoError(t, d.DB().Model(&latest).Update("is_pending", false).Error)
	updated, err := d.GetOverview("A", 7, now)
	require.NoError(t, err)
	require.EqualValues(t, 4, updated.Trend[6].Published)
	require.Zero(t, updated.Trend[6].Pending)
	require.EqualValues(t, 3, updated.Metrics.Participants)

	// A soft-deleted user is no longer an existing participant.
	require.NoError(t, d.DB().Delete(&admin).Error)
	updated, err = d.GetOverview("A", 7, now)
	require.NoError(t, err)
	require.EqualValues(t, 2, updated.Metrics.Participants)
}

func TestOverviewCalendarBoundaries(t *testing.T) {
	for _, zone := range []string{"Asia/Shanghai", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			d := overviewDAO(t)
			location, err := time.LoadLocation(zone)
			require.NoError(t, err)
			d.SetLocation(location)
			// New York's spring transition is a 23-hour calendar day.
			today := time.Date(2026, 3, 8, 0, 0, 0, 0, location)
			end := today.AddDate(0, 0, 1)
			for _, at := range []time.Time{today.AddDate(0, 0, -6).Add(-time.Second), today.AddDate(0, 0, -6), today.Add(-time.Second), today, end.Add(-time.Second), end} {
				c := entity.Comment{Model: gorm.Model{CreatedAt: at.In(time.Local)}, Content: "boundary", SiteName: "A"}
				require.NoError(t, d.DB().Create(&c).Error)
			}
			result, err := d.GetOverview("A", 7, today.Add(12*time.Hour))
			require.NoError(t, err)
			require.EqualValues(t, 2, result.Metrics.TodayComments)
			require.EqualValues(t, 1, result.Trend[0].Published)
			require.EqualValues(t, 1, result.Trend[5].Published)
			require.EqualValues(t, 2, result.Trend[6].Published)
		})
	}
}

func TestOverviewEmpty(t *testing.T) {
	d := overviewDAO(t)
	result, err := d.GetOverview("", 7, time.Now())
	require.NoError(t, err)
	require.Equal(t, dao.OverviewMetrics{}, result.Metrics)
	require.Len(t, result.Trend, 7)
	require.Empty(t, result.LatestComments)
	require.NotNil(t, result.LatestComments)
	require.Empty(t, result.PopularPages)
	require.NotNil(t, result.PopularPages)
	_, err = d.GetOverview("", 365, time.Now())
	require.Error(t, err)
}
