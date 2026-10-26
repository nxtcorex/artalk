package handler

import (
	"strconv"
	"time"

	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/gofiber/fiber/v2"
)

type ResponseCommentActivity = dao.CommentActivity

// @Id           GetCommentActivity
// @Summary      Comment activity calendar
// @Description  Daily new comments, including published and pending comments, in the configured timezone. Omit year for the last 365 days. Available years follow the first non-deleted comment in the selected site scope.
// @Tags         Admin
// @Security     ApiKeyAuth
// @Param        site_name query string false "Site name; empty includes all sites"
// @Param        year query int false "Calendar year, from 1970 through the current year"
// @Produce      json
// @Success      200 {object} ResponseCommentActivity
// @Failure      400 {object} Map{msg=string}
// @Failure      403 {object} Map{msg=string}
// @Failure      404 {object} Map{msg=string}
// @Failure      500 {object} Map{msg=string}
// @Router       /admin/overview/activity [get]
func CommentActivity(app *core.App, router fiber.Router) {
	router.Get("/admin/overview/activity", common.AdminGuard(app, func(c *fiber.Ctx) error {
		now := time.Now()
		year := 0
		if value := c.Query("year"); value != "" {
			var err error
			year, err = strconv.Atoi(value)
			if err != nil || year < 1970 || year > now.In(app.Dao().Location()).Year() {
				return common.RespError(c, 400, "Invalid activity year")
			}
		}
		siteName := c.Query("site_name")
		if siteName != "" {
			if _, ok, resp := common.CheckSiteExist(app, c, siteName); !ok {
				return resp
			}
		}
		result, err := app.Dao().GetCommentActivity(siteName, year, now)
		if err != nil {
			return common.RespError(c, 500, "Failed to load comment activity")
		}
		return common.RespData(c, result)
	}))
}
