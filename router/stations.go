package router

import (
	"github.com/memphisdev/memphis-rest-gateway/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitializeStationsRoutes(app *fiber.App) {
	api := app.Group("/stations", logger.New())
	api.Post("/:stationName/produce/single", handlers.CreateHandleMessage())
	api.Post("/:stationName/produce/batch", handlers.CreateHandleBatch())
	api.Post("/:stationName/consume/batch", handlers.ConsumeHandleMessage())
}
