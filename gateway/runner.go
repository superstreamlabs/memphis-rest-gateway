package gateway

import (
	"fmt"

	"github.com/memphisdev/memphis-rest-gateway/conf"
	"github.com/memphisdev/memphis-rest-gateway/handlers"
	"github.com/memphisdev/memphis-rest-gateway/logger"
	mconntr "github.com/memphisdev/memphis-rest-gateway/memphisSingleton"
	"github.com/memphisdev/memphis-rest-gateway/router"
	"github.com/nats-io/nats.go"
)

func Run(cnf conf.Configuration, lgr *logger.Logger, conn *nats.Conn) error {
	mconntr.Put(conn)

	err := handlers.ListenForUpdates(lgr)
	if err != nil {
		return fmt.Errorf("Error while listening for updates - %s", err.Error())
	}
	go handlers.CleanConnectionsCache()
	app := router.SetupRoutes(lgr)
	lgr.Noticef("Memphis REST gateway is up and running")
	lgr.Noticef("Version %s", cnf.VERSION)
	return app.Listen(":" + cnf.HTTP_PORT)
}

/*
func initializeLogger() *logger.Logger {
	configuration := conf.Get()
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			creds := configuration.CONNECTION_TOKEN
			username := configuration.ROOT_USER
			if configuration.USER_PASS_BASED_AUTH {
				username = "$$memphis"
				creds = configuration.CONNECTION_TOKEN + "_" + configuration.ROOT_PASSWORD
				if !configuration.CLOUD_ENV {
					creds = configuration.ROOT_PASSWORD
				}
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
	configuration := conf.Get()
	l := initializeLogger()
	err := handlers.ListenForUpdates(l)
	if err != nil {
		panic("Error while listening for updates - " + err.Error())
	}
	go handlers.CleanConnectionsCache()
	app := router.SetupRoutes(l)
	l.Noticef("Memphis REST gateway is up and running")
	l.Noticef("Version %s", configuration.VERSION)
	app.Listen(":" + configuration.HTTP_PORT)
}
*/
