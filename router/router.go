package router

import (
	"github.com/memphisdev/memphis-rest-gateway/logger"
	"github.com/memphisdev/memphis-rest-gateway/middlewares"
	"github.com/memphisdev/memphis-rest-gateway/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupRoutes setup router api
func SetupRoutes(l *logger.Logger) *fiber.App {
	utils.InitializeValidations()
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	logger.SetLogger(app, l)
	app.Use(cors.New())
	app.Use(middlewares.Authenticate)

	InitilizeAuthRoutes(app)
	InitializeStationsRoutes(app)
	InitilizeMonitoringRoutes(app)
	return app
}
