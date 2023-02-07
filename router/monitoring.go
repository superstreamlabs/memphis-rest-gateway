package router

import (
	"rest-gateway/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitilizeMonitoringRoutes(app *fiber.App) {
	monitoringHandlerHandler := handlers.MonitoringHandler{}
	api := app.Group("/monitoring", logger.New())
	api.Get("/status", monitoringHandlerHandler.Status)
	api.Get("/getResourcesUtilization", monitoringHandlerHandler.GetResourcesUtilization)
}
