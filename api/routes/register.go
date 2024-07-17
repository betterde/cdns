package routes

import (
	"github.com/betterde/cdns/api/handler"
	"github.com/betterde/cdns/internal/response"
	"github.com/betterde/cdns/spa"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(response.Success("Success", nil))
	}).Name("Health check")

	app.Post("/present", handler.Present).Name("Create TXT record")
	app.Post("/cleanup", handler.Cleanup).Name("Cleanup TXT record")

	// Embed SPA static resource
	app.Get("*", filesystem.New(filesystem.Config{
		Root:               spa.Serve(),
		Index:              "index.html",
		NotFoundFile:       "index.html",
		ContentTypeCharset: "UTF-8",
	})).Name("SPA static resource")
}
