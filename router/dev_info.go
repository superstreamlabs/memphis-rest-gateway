package router

import (
	"http-proxy/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitilizeDevInfoRoutes(app *fiber.App) {
	devInfoHandler := handlers.DevInfoHandler{}
	api := app.Group("/dev", logger.New())
	api.Get("/getSystemInfo", devInfoHandler.GetSystemInfo)
}
