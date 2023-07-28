package main

import (
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/handlers"
	"rest-gateway/logger"
	"rest-gateway/router"
	"time"
)

func initalizeLogger() *logger.Logger {
	configuration := conf.GetConfig()
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			creds := configuration.CONNECTION_TOKEN
			username := configuration.ROOT_USER
			if configuration.USER_PASS_BASED_AUTH {
				username = "$memphis"
				creds = configuration.CONNECTION_TOKEN + "_" + configuration.ROOT_PASSWORD
			}
			l, err := logger.CreateLogger(configuration.MEMPHIS_HOST, username, creds)
			if err != nil {
				fmt.Printf("Awaiting to establish connection with Memphis - %v\n", err.Error())
			} else {
				ticker.Stop()
				return l
			}
		}
	}
}

func main() {
	configuration := conf.GetConfig()
	l := initalizeLogger()
	go handlers.CleanConnectionsCache()
	app := router.SetupRoutes(l)
	l.Noticef("Memphis REST gateway is up and running")
	l.Noticef("Version %s", configuration.VERSION)
	app.Listen(":" + configuration.HTTP_PORT)

}
