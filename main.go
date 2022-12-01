package main

import (
	"fmt"
	"http-proxy/conf"
	"http-proxy/router"
	"time"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/memphisdev/memphis.go"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())

	time.Sleep(5 * time.Second)
	configuration := conf.GetConfig()
	conn, err := memphis.Connect(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	router.SetupRoutes(app, conn)
	app.Listen(":" + configuration.HTTP_PORT)
}
