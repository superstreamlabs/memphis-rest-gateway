package router

import (
	"rest-gateway/logger"
	"rest-gateway/middlewares"
	"rest-gateway/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/memphisdev/memphis.go"
)

// SetupRoutes setup router api
func SetupRoutes(conn *memphis.Conn, l *logger.Logger) *fiber.App {
	utils.InitializeValidations()
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	logger.SetLogger(app, l)
	app.Use(cors.New())
	app.Use(middlewares.Authenticate)

	InitilizeAuthRoutes(app)
	InitializeStationsRoutes(app, conn)
	InitilizeMonitoringRoutes(app)
	return app
}
