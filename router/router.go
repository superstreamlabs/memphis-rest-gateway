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

	// Flag to track if a specific route has been matched
	routeMatched := false

	logger.SetLogger(app, l)
	app.Use(cors.New())

	app.Post("/stations/:stationName/produce/single", func(c *fiber.Ctx) error {
		return middlewares.AuthenticateStation(c, &routeMatched)
	})
	app.Post("/stations/:stationName/produce/batch", func(c *fiber.Ctx) error {
		return middlewares.AuthenticateStation(c, &routeMatched)
	})

	InitilizeAuthRoutes(app)

	// Register default middleware for unmatched routes. The order is important, this has to be called after InitilizeAuthRoutes().
	app.Use(func(c *fiber.Ctx) error {
		// Check if a specific route has been matched
		if !routeMatched {
			// Apply default middleware logic for unmatched routes
			return middlewares.Authenticate(c)
		}

		// Reset the routeMatched flag for subsequent requests
		routeMatched = false

		// Continue processing the request
		return c.Next()
	})

	InitializeStationsRoutes(app, conn)
	InitilizeMonitoringRoutes(app)

	return app
}
