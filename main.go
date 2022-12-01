package main

import (
	"fmt"
	"http-proxy/conf"
	"http-proxy/router"

	"os"

	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

func main() {
	configuration := conf.GetConfig()
	conn, err := memphis.Connect(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	app := router.SetupRoutes(conn)
	app.Listen(configuration.HTTP_PORT)
}
