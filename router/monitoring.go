package router

import (
	"http-proxy/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitilizeMonitoringRoutes(app *fiber.App) {
	devInfoHandler := handlers.DevInfoHandler{}
	api := app.Group("/monitoring", logger.New())
	api.Get("/getResourcesUtilization", devInfoHandler.GetResourcesUtilization)
}
