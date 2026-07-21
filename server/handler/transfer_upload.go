package handler

import (
	"errors"

	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/artalkjs/artalk/v2/internal/log"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/gofiber/fiber/v2"
)

type ResponseTransferUpload struct {
	// The uploaded file name which can be used to import
	Filename string `json:"filename"`
}

// @Id           UploadArtrans
// @Summary      Upload Artrans
// @Description  Upload a file to prepare to import
// @Tags         Transfer
// @Security     ApiKeyAuth
// @Param        file  formData  file  true  "Upload file in preparation for import task"
// @Accept       mpfd
// @Produce      json
// @Success      200  {object}  ResponseTransferUpload{filename=string}
// @Failure      500  {object}  Map{msg=string}
// @Router       /transfer/upload  [post]
func TransferUpload(app *core.App, router fiber.Router, files *transferFileStore) {
	router.Post("/transfer/upload", common.AdminGuard(app, func(c *fiber.Ctx) error {
		// Get file from FormData
		file, err := c.FormFile("file")
		if err != nil {
			log.Error(err)
			return common.RespError(c, 500, "File read failed")
		}

		// Open file
		src, err := file.Open()
		if err != nil {
			log.Error(err)
			return common.RespError(c, 500, "File open failed")
		}
		defer src.Close()

		maxBytes := int64(app.Conf().HTTP.BodyLimit) * 1024 * 1024
		filename, err := files.Save(src, maxBytes)
		if err != nil {
			log.Error(err)
			if errors.Is(err, errTransferFileTooLarge) {
				return common.RespError(c, fiber.StatusRequestEntityTooLarge, "File is too large")
			}
			return common.RespError(c, fiber.StatusInternalServerError, "File save failed")
		}

		// Return filename
		return common.RespData(c, ResponseTransferUpload{
			Filename: filename,
		})
	}))
}
