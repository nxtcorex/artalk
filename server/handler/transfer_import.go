package handler

import (
	"bytes"
	"html"
	"os"
	"sync"

	"github.com/artalkjs/artalk/v2/internal/artransfer"
	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/gofiber/fiber/v2"
)

type ParamsTransferImport struct {
	artransfer.ImportParams
}

// @Id           ImportArtrans
// @Summary      Import Artrans
// @Description  Import data to Artalk
// @Tags         Transfer
// @Security     ApiKeyAuth
// @Param        data  body  ParamsTransferImport  true  "The data to import"
// @Accept       json
// @Produce      html
// @Success      200  {string}  string
// @Failure      400  {object}  Map{msg=string}
// @Failure      422  {string}  string
// @Router       /transfer/import  [post]
func TransferImport(app *core.App, router fiber.Router, files *transferFileStore) {
	var mu sync.Mutex

	router.Post("/transfer/import", common.AdminGuard(app, func(c *fiber.Ctx) error {
		if !mu.TryLock() {
			return common.RespError(c, fiber.StatusTooManyRequests, "Another import is in progress")
		}
		defer mu.Unlock()

		var p ParamsTransferImport
		if isOK, resp := common.ParamsDecode(c, &p); !isOK {
			files.Discard(c.FormValue("json_file"))
			return resp
		}

		if p.JsonFile != "" {
			filename, ok := files.Take(p.JsonFile)
			if !ok {
				return common.RespError(c, fiber.StatusBadRequest, "Import file is missing, expired, or already consumed")
			}
			p.JsonFile = filename
			defer os.Remove(filename)
		}

		buf := bytes.NewBufferString("<style>* { font-family: Menlo, Consolas, Monaco, monospace;word-wrap: break-word;white-space: pre-wrap;font-size: 13px; }</style>")

		p.Assumeyes = true
		err := artransfer.RunImportArtrans(app.Dao(), &p.ImportParams, func(s string) {
			buf.WriteString(html.EscapeString(s))
		})
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).Type("html", "utf-8").SendString(buf.String())
		}

		return c.Status(fiber.StatusOK).Type("html", "utf-8").SendString(buf.String())
	}))
}
