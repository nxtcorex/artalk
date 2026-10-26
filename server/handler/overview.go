package handler

import (
	"time"

	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/gofiber/fiber/v2"
)

type ResponseOverview = dao.Overview

// @Id           GetOverview
// @Summary      Admin overview
// @Description  Comment activity in the configured timezone. Historical buckets use current moderation status. Participants are distinct existing users with published comments in the selected period.
// @Tags         Admin
// @Security     ApiKeyAuth
// @Param        site_name query string false "Site name; empty includes all sites"
// @Param        days query int false "Calendar days including today" Enums(7, 30) default(7)
// @Produce      json
// @Success      200 {object} ResponseOverview
// @Failure      400 {object} Map{msg=string}
// @Failure      403 {object} Map{msg=string}
// @Failure      404 {object} Map{msg=string}
// @Failure      500 {object} Map{msg=string}
// @Router       /admin/overview [get]
func Overview(app *core.App, router fiber.Router) {
	router.Get("/admin/overview", common.AdminGuard(app, func(c *fiber.Ctx) error {
		days := 7
		switch c.Query("days", "7") {
		case "7":
		case "30":
			days = 30
		default:
			return common.RespError(c, 400, "days must be 7 or 30")
		}
		siteName := c.Query("site_name")
		if siteName != "" {
			if _, ok, resp := common.CheckSiteExist(app, c, siteName); !ok {
				return resp
			}
		}
		result, err := app.Dao().GetOverview(siteName, days, time.Now())
		if err != nil {
			return common.RespError(c, 500, "Failed to load overview")
		}
		return common.RespData(c, result)
	}))
}
