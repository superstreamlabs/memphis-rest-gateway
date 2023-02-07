package main

import (
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/router"
	"time"

	"github.com/memphisdev/memphis.go"
)

func main() {
	configuration := conf.GetConfig()
	var conn *memphis.Conn
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			var err error
			if configuration.CLIENT_CERT_PATH != "" && configuration.CLIENT_KEY_PATH != "" && configuration.ROOT_CA_PATH != "" {
				conn, err = memphis.Connect(
					configuration.MEMPHIS_HOST,
					configuration.ROOT_USER,
					configuration.CONNECTION_TOKEN,
					memphis.Reconnect(true),
					memphis.MaxReconnect(10),
					memphis.ReconnectInterval(3*time.Second),
					memphis.Tls(configuration.CLIENT_CERT_PATH, configuration.CLIENT_KEY_PATH, configuration.ROOT_CA_PATH),
				)
			} else {
				conn, err = memphis.Connect(
					configuration.MEMPHIS_HOST,
					configuration.ROOT_USER,
					configuration.CONNECTION_TOKEN,
					memphis.Reconnect(true),
					memphis.MaxReconnect(10),
					memphis.ReconnectInterval(3*time.Second),
				)
			}
			if err == nil {
				ticker.Stop()
				goto serverInit
			} else {
				fmt.Printf("Awaiting to establish connection with Memphis - %v\n", err.Error())
			}
		}
	}

serverInit:
	l, err := logger.CreateLogger(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
	if err != nil {
		panic("Logger creation failed - " + err.Error())
	}

	app := router.SetupRoutes(conn, l)
	l.Noticef("Memphis REST gateway is up and running")
	l.Noticef("Version %s", configuration.VERSION)
	app.Listen(":" + configuration.HTTP_PORT)
}
