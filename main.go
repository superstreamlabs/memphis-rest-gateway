package main

import (
	"fmt"
	"http-proxy/conf"
	"http-proxy/router"
	"log"
	"time"

	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

func main() {
	configuration := conf.GetConfig()
	var conn *memphis.Conn
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			var err error
			conn, err = memphis.Connect(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
			if err == nil {
				ticker.Stop()
				goto serverInit
			} else {
				fmt.Printf("Awaiting to establish connection with Memphis - %v", err.Error())
			}
		}
	}

serverInit:
	app := router.SetupRoutes(conn)
	log.Output(1, "Memphis Http Proxy is up and running")
	app.Listen(":" + configuration.HTTP_PORT)
}
