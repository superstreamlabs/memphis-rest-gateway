package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/memphisdev/memphis.go"
	handler "http-proxy/handlers"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App, conn *memphis.Conn) {
	api := app.Group("/", logger.New())
	api.Post("/station/:stationName/:producerName", handler.CreateHandleMessage(conn))
}
