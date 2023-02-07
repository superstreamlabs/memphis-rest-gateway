package router

import (
	"rest-gateway/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/memphisdev/memphis.go"
)

func InitializeStationsRoutes(app *fiber.App, conn *memphis.Conn) {
	api := app.Group("/stations", logger.New())
	api.Post("/:stationName/produce/single", handlers.CreateHandleMessage(conn))
	api.Post("/:stationName/produce/batch", handlers.CreateHandleBatch(conn))
}
