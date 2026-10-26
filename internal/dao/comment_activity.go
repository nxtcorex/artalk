package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/artalkjs/artalk/v2/internal/entity"
)

type CommentActivity struct {
	Timezone    string               `json:"timezone"`
	GeneratedAt string               `json:"generated_at"`
	StartDate   string               `json:"start_date"`
	EndDate     string               `json:"end_date"`
	Years       []int                `json:"years"`
	Days        []CommentActivityDay `json:"days"`
}

type CommentActivityDay struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetCommentActivity counts published and pending comments by creation date.
// A zero year selects the last 365 days; a calendar year includes its entire
// grid, with future dates left empty. Soft-deleted comments are excluded.
func (dao *Dao) GetCommentActivity(siteName string, year int, now time.Time) (CommentActivity, error) {
	location := dao.Location()
	now = now.In(location)
	if year != 0 && (year < 1970 || year > now.Year()) {
		return CommentActivity{}, fmt.Errorf("invalid activity year")
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	start, end := today.AddDate(0, 0, -364), today.AddDate(0, 0, 1)
	if year != 0 {
		start = time.Date(year, 1, 1, 0, 0, 0, 0, location)
		end = start.AddDate(1, 0, 0)
	}
	queryEnd := end
	if queryEnd.After(today.AddDate(0, 0, 1)) {
		queryEnd = today.AddDate(0, 0, 1)
	}
	result := CommentActivity{
		Timezone: location.String(), GeneratedAt: now.Format(time.RFC3339),
		StartDate: start.Format("2006-01-02"), EndDate: end.AddDate(0, 0, -1).Format("2006-01-02"),
		Years: []int{}, Days: []CommentActivityDay{},
	}
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		result.Days = append(result.Days, CommentActivityDay{Date: day.Format("2006-01-02")})
	}
	comments := func() *DB {
		q := dao.db.Model(&entity.Comment{})
		if siteName != "" {
			q = q.Where("site_name = ?", siteName)
		}
		return q
	}
	var first entity.Comment
	if err := comments().Select("created_at").Where("created_at < ?", today.AddDate(0, 0, 1).In(time.Local)).
		Order("created_at ASC").Limit(1).Find(&first).Error; err != nil {
		return result, err
	}
	firstYear := now.Year()
	if !first.CreatedAt.IsZero() {
		firstYear = max(1970, first.CreatedAt.In(location).Year())
	}
	for y := now.Year(); y >= firstYear; y-- {
		result.Years = append(result.Years, y)
	}

	// Group a derived calendar bucket instead of using SQL date functions. The
	// boundaries honor the configured timezone and DST on every SQL dialect.
	var cases []string
	var args []any
	for i, day := 0, start; day.Before(queryEnd); i, day = i+1, day.AddDate(0, 0, 1) {
		cases = append(cases, fmt.Sprintf("WHEN created_at < ? THEN %d", i))
		args = append(args, day.AddDate(0, 0, 1).In(time.Local))
	}
	buckets := comments().Where("created_at >= ? AND created_at < ?", start.In(time.Local), queryEnd.In(time.Local)).
		Select("CASE "+strings.Join(cases, " ")+" END AS bucket", args...)
	var counts []struct {
		Bucket int
		Count  int64
	}
	if err := dao.db.Table("(?) AS activity", buckets).Select("bucket, COUNT(*) AS count").Group("bucket").Scan(&counts).Error; err != nil {
		return result, err
	}
	for _, count := range counts {
		result.Days[count.Bucket].Count = count.Count
	}
	return result, nil
}
