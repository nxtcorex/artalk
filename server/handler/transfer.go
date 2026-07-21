package handler

import (
	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/gofiber/fiber/v2"
)

func Transfer(app *core.App, router fiber.Router) {
	files := newTransferFileStore(transferFileTTL)
	app.OnTerminate().Add(func(_ *core.TerminateEvent) error {
		files.Close()
		return nil
	})

	TransferImport(app, router, files)
	TransferUpload(app, router, files)
	TransferExport(app, router)
}
