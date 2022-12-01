package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/memphisdev/memphis.go"
	handler "http-proxy/handlers"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App, conn *memphis.Conn) {
	api := app.Group("/stations", logger.New())
	api.Post("/:stationName/produce/single", handler.CreateHandleMessage(conn))
}
