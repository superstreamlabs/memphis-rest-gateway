package router

import (
	"rest-gateway/handlers"
	"rest-gateway/logger"
	"rest-gateway/middlewares"
	"rest-gateway/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupRoutes setup router api
func SetupRoutes(l *logger.Logger) *fiber.App {
	utils.InitializeValidations()
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	handlers.InitializeMessageCache(l)

	logger.SetLogger(app, l)
	app.Use(cors.New())
	app.Use(middlewares.Authenticate)

	InitilizeAuthRoutes(app)
	InitializeStationsRoutes(app)
	InitilizeMonitoringRoutes(app)
	return app
}
