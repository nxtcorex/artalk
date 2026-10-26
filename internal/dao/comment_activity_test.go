package dao_test

import (
	"testing"
	"time"

	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCommentActivity(t *testing.T) {
	d := overviewDAO(t)
	location, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	d.SetLocation(location)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, location)
	today := time.Date(2026, 10, 26, 0, 0, 0, 0, location)
	add := func(site string, at time.Time, pending bool) entity.Comment {
		c := entity.Comment{Model: gorm.Model{CreatedAt: at.In(time.Local)}, SiteName: site, Content: "activity", IsPending: pending}
		require.NoError(t, d.DB().Create(&c).Error)
		return c
	}
	date := func(year, month, day int) time.Time {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, location)
	}
	add("A", date(2023, 12, 31), false)
	add("A", date(2024, 1, 1), false)
	add("A", date(2024, 2, 29), false)
	add("A", date(2024, 2, 29), true)
	add("B", date(2024, 2, 29), false)
	add("A", date(2025, 1, 1).Add(-time.Second), false)
	deleted := add("A", date(2024, 2, 29), false)
	require.NoError(t, d.DB().Delete(&deleted).Error)
	add("A", today.AddDate(0, 0, -364), false)
	add("A", today.AddDate(0, 0, -364).Add(-time.Second), false)
	add("A", now, true)
	add("B", now, false)
	add("A", today.AddDate(0, 0, 1), false)

	year, err := d.GetCommentActivity("A", 2024, now)
	require.NoError(t, err)
	require.Equal(t, "Asia/Shanghai", year.Timezone)
	require.Equal(t, now.Format(time.RFC3339), year.GeneratedAt)
	require.Equal(t, "2024-01-01", year.StartDate)
	require.Equal(t, "2024-12-31", year.EndDate)
	require.Equal(t, []int{2026, 2025, 2024, 2023}, year.Years)
	require.Len(t, year.Days, 366)
	require.EqualValues(t, 1, year.Days[0].Count)
	require.Equal(t, "2024-02-29", year.Days[59].Date)
	require.EqualValues(t, 2, year.Days[59].Count)
	require.Zero(t, year.Days[60].Count)
	require.EqualValues(t, 1, year.Days[365].Count)

	rolling, err := d.GetCommentActivity("A", 0, now)
	require.NoError(t, err)
	require.Len(t, rolling.Days, 365)
	require.Equal(t, "2025-10-27", rolling.StartDate)
	require.Equal(t, "2026-10-26", rolling.EndDate)
	var total int64
	for _, day := range rolling.Days {
		total += day.Count
	}
	require.EqualValues(t, 2, total)
	require.EqualValues(t, 1, rolling.Days[0].Count)
	require.EqualValues(t, 1, rolling.Days[364].Count)
	all, err := d.GetCommentActivity("", 2024, now)
	require.NoError(t, err)
	require.EqualValues(t, 3, all.Days[59].Count)
	current, err := d.GetCommentActivity("A", 2026, now)
	require.NoError(t, err)
	require.Len(t, current.Days, 365)
	for _, day := range current.Days {
		if day.Date > "2026-10-26" {
			require.Zero(t, day.Count, day.Date)
		}
	}
}

func TestCommentActivityCalendarBoundaries(t *testing.T) {
	for _, zone := range []string{"Asia/Shanghai", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			d := overviewDAO(t)
			location, err := time.LoadLocation(zone)
			require.NoError(t, err)
			d.SetLocation(location)
			today := time.Date(2026, 3, 8, 0, 0, 0, 0, location)
			end := today.AddDate(0, 0, 1)
			for _, at := range []time.Time{today.Add(-time.Second), today, end.Add(-time.Second), end} {
				c := entity.Comment{Model: gorm.Model{CreatedAt: at.In(time.Local)}, Content: "boundary"}
				require.NoError(t, d.DB().Create(&c).Error)
			}
			result, err := d.GetCommentActivity("", 0, today.Add(12*time.Hour))
			require.NoError(t, err)
			require.EqualValues(t, 1, result.Days[363].Count)
			require.EqualValues(t, 2, result.Days[364].Count)
		})
	}
}

func TestCommentActivityEmpty(t *testing.T) {
	d := overviewDAO(t)
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	result, err := d.GetCommentActivity("", 0, now)
	require.NoError(t, err)
	require.Equal(t, []int{2026}, result.Years)
	require.Len(t, result.Days, 365)
	for _, day := range result.Days {
		require.Zero(t, day.Count)
	}
	for _, year := range []int{-1, 1969, 2027} {
		_, err := d.GetCommentActivity("", year, now)
		require.Error(t, err)
	}
}
