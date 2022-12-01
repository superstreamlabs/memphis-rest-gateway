package router

import (
	handler "http-proxy/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/memphisdev/memphis.go"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App, conn *memphis.Conn) {
	api := app.Group("/stations", logger.New())
	api.Post("/:stationName/produce/single", handler.CreateHandleMessage(conn))
	api.Post("/:stationName/produce/batch", handler.CreateHandleBatch(conn))
}
