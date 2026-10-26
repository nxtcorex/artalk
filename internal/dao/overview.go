package dao

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/artalkjs/artalk/v2/internal/utils"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Overview struct {
	Timezone       string            `json:"timezone"`
	GeneratedAt    string            `json:"generated_at"`
	Metrics        OverviewMetrics   `json:"metrics"`
	Trend          []OverviewDay     `json:"trend"`
	LatestComments []OverviewComment `json:"latest_comments"`
	PopularPages   []OverviewPage    `json:"popular_pages"`
}

type OverviewMetrics struct {
	TodayComments  int64 `json:"today_comments"`
	Pending        int64 `json:"pending"`
	Published      int64 `json:"published"`
	Participants   int64 `json:"participants"`
	CommentedPages int64 `json:"commented_pages"`
	PV             int64 `json:"pv"`
}

type OverviewDay struct {
	Date      string `json:"date"`
	Published int64  `json:"published"`
	Pending   int64  `json:"pending"`
}

type OverviewComment struct {
	ID        uint   `json:"id"`
	Nick      string `json:"nick"`
	Summary   string `json:"summary"`
	PageTitle string `json:"page_title"`
	PageKey   string `json:"page_key"`
	PageURL   string `json:"page_url"`
	SiteName  string `json:"site_name"`
	CreatedAt string `json:"created_at"`
	IsPending bool   `json:"is_pending"`
}

type OverviewPage struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Key      string `json:"key"`
	URL      string `json:"url"`
	SiteName string `json:"site_name"`
	Comments int64  `json:"comments"`
}

// GetOverview aggregates in SQL without dialect-specific date functions. Each
// calendar boundary is calculated in the configured timezone, including DST.
func (dao *Dao) GetOverview(siteName string, days int, now time.Time) (Overview, error) {
	if days != 7 && days != 30 {
		return Overview{}, fmt.Errorf("overview days must be 7 or 30")
	}
	location := dao.Location()
	now = now.In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	start := today.AddDate(0, 0, 1-days)
	end := today.AddDate(0, 0, 1).In(time.Local)
	result := Overview{
		Timezone: location.String(), GeneratedAt: now.Format(time.RFC3339),
		Trend: make([]OverviewDay, days), LatestComments: []OverviewComment{}, PopularPages: []OverviewPage{},
	}
	scope := func(q *gorm.DB) *gorm.DB {
		if siteName != "" {
			return q.Where("site_name = ?", siteName)
		}
		return q
	}
	comments := func() *gorm.DB { return dao.db.Model(&entity.Comment{}).Scopes(scope) }
	if err := comments().Select(
		"COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN 1 ELSE 0 END), 0) AS today_comments, "+
			"COALESCE(SUM(CASE WHEN is_pending = ? THEN 1 ELSE 0 END), 0) AS pending, "+
			"COALESCE(SUM(CASE WHEN is_pending = ? THEN 1 ELSE 0 END), 0) AS published",
		today.In(time.Local), end, true, false,
	).Scan(&result.Metrics).Error; err != nil {
		return result, err
	}
	if err := comments().Where("is_pending = ? AND created_at >= ? AND created_at < ?", false, start.In(time.Local), end).
		Where("user_id IN (?)", dao.db.Model(&entity.User{}).Select("id")).
		Distinct("user_id").Count(&result.Metrics.Participants).Error; err != nil {
		return result, err
	}

	// A bounded projection returns at most 60 scalar values, never comment rows.
	selects := make([]string, 0, days*2)
	args := make([]any, 0, days*6)
	values := make([]any, 0, days*2)
	for i := range days {
		day := start.AddDate(0, 0, i)
		result.Trend[i].Date = day.Format("2006-01-02")
		for _, pending := range []bool{false, true} {
			selects = append(selects, "COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? AND is_pending = ? THEN 1 ELSE 0 END), 0)")
			args = append(args, day.In(time.Local), day.AddDate(0, 0, 1).In(time.Local), pending)
		}
		values = append(values, &result.Trend[i].Published, &result.Trend[i].Pending)
	}
	row := comments().Where("created_at >= ? AND created_at < ?", start.In(time.Local), end).
		Select(strings.Join(selects, ", "), args...).Row()
	if err := row.Scan(values...); err != nil {
		return result, err
	}

	pagesTable := dao.GetTableName(&entity.Page{})
	commentsTable := dao.GetTableName(&entity.Comment{})
	relatedComments := func() *gorm.DB {
		return dao.db.Model(&entity.Comment{}).
			Where(clause.Eq{Column: clause.Column{Table: commentsTable, Name: "page_key"}, Value: clause.Column{Table: pagesTable, Name: "key"}}).
			Where(clause.Eq{Column: clause.Column{Table: commentsTable, Name: "site_name"}, Value: clause.Column{Table: pagesTable, Name: "site_name"}})
	}
	if err := dao.db.Model(&entity.Page{}).Scopes(scope).Where("EXISTS (?)", relatedComments().Select("1")).
		Count(&result.Metrics.CommentedPages).Error; err != nil {
		return result, err
	}
	if err := dao.db.Model(&entity.Page{}).Scopes(scope).Select("COALESCE(SUM(pv), 0)").Scan(&result.Metrics.PV).Error; err != nil {
		return result, err
	}

	var latest []entity.Comment
	if err := comments().Preload("User").Preload("Page.Site").Order("created_at DESC").Order("id DESC").Limit(5).Find(&latest).Error; err != nil {
		return result, err
	}
	for _, comment := range latest {
		page := comment.Page
		if page == nil {
			page = &entity.Page{Key: comment.PageKey, SiteName: comment.SiteName}
		}
		nick := ""
		if comment.User != nil {
			nick = comment.User.Name
		}
		result.LatestComments = append(result.LatestComments, OverviewComment{
			ID: comment.ID, Nick: nick, Summary: overviewSummary(comment.Content),
			PageTitle: page.Title, PageKey: comment.PageKey, PageURL: dao.GetPageAccessibleURL(page, page.Site),
			SiteName: comment.SiteName, CreatedAt: comment.CreatedAt.In(location).Format(time.RFC3339), IsPending: comment.IsPending,
		})
	}

	// Correlate on both keys. Limiting page IDs first avoids GROUP BY rules for
	// page metadata that differ across PostgreSQL, MySQL and SQL Server.
	var popular []struct {
		ID       uint
		Comments int64
	}
	count := relatedComments().Select("COUNT(*)").Where("is_pending = ? AND created_at >= ? AND created_at < ?", false, start.In(time.Local), end)
	if err := dao.db.Model(&entity.Page{}).Scopes(scope).
		Select("id, (?) AS comments", count).Where("EXISTS (?)", relatedComments().Select("1").Where("is_pending = ? AND created_at >= ? AND created_at < ?", false, start.In(time.Local), end)).
		Order("comments DESC").Order("id ASC").Limit(5).Scan(&popular).Error; err != nil {
		return result, err
	}
	if len(popular) > 0 {
		ids := make([]uint, 0, len(popular))
		for _, page := range popular {
			ids = append(ids, page.ID)
		}
		var pages []entity.Page
		if err := dao.db.Preload("Site").Find(&pages, ids).Error; err != nil {
			return result, err
		}
		byID := make(map[uint]entity.Page, len(pages))
		for _, page := range pages {
			byID[page.ID] = page
		}
		for _, item := range popular {
			page, ok := byID[item.ID]
			if !ok {
				continue
			}
			result.PopularPages = append(result.PopularPages, OverviewPage{
				ID: page.ID, Title: page.Title, Key: page.Key, URL: dao.GetPageAccessibleURL(&page, page.Site),
				SiteName: page.SiteName, Comments: item.Comments,
			})
		}
	}
	return result, nil
}

var overviewTextPolicy = bluemonday.StrictPolicy()

func overviewSummary(content string) string {
	marked, _ := utils.Marked(content)
	text := strings.Join(strings.Fields(html.UnescapeString(overviewTextPolicy.Sanitize(marked))), " ")
	runes := []rune(text)
	if len(runes) > 180 {
		return string(runes[:180]) + "…"
	}
	return text
}
