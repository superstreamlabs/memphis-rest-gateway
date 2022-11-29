package main

import (
	"fmt"
	"http-proxy/router"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/memphisdev/memphis.go"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())

	conn, err := memphis.Connect("localhost", "root", "memphis")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	producer, err := conn.CreateProducer("test-fiber-go", "simple_go_producer")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	router.SetupRoutes(app, producer)
	app.Listen(":3000")
}
