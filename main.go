package main

import (
	"os"
	"os/signal"
	"rest-gateway/conf"
	"rest-gateway/handlers"
	"rest-gateway/logger"
	"rest-gateway/router"
	"syscall"
	"time"
)

func initalizeLogger() {
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
				panic("Logger creation failed - " + err.Error())
			}

			app := router.SetupRoutes(l)
			l.Noticef("Memphis REST gateway is up and running")
			l.Noticef("Version %s", configuration.VERSION)
			app.Listen(":" + configuration.HTTP_PORT)
		}
	}
}

func main() {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)
	go initalizeLogger()
	go handlers.CleanConnectionsCache()
	<-interruptCh
}
