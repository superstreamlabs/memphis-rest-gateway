package router

import (
	"github.com/memphisdev/memphis-rest-gateway/conf"
	"github.com/memphisdev/memphis-rest-gateway/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitilizeMonitoringRoutes(app *fiber.App) {
	configuration := conf.Get()

	monitoringHandlerHandler := handlers.MonitoringHandler{}
	api := app.Group("/monitoring", logger.New())
	api.Get("/status", monitoringHandlerHandler.Status)
	if configuration.DEV_ENV == "true" {
		api.Get("/getResourcesUtilization", monitoringHandlerHandler.GetResourcesUtilization)
	}
}
