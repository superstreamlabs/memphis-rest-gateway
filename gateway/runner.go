package gateway

import (
	"fmt"
	"time"

	"github.com/g41797/sputnik"
	"github.com/memphisdev/memphis-rest-gateway/conf"
	"github.com/memphisdev/memphis-rest-gateway/handlers"
	"github.com/memphisdev/memphis-rest-gateway/logger"
	mconntr "github.com/memphisdev/memphis-rest-gateway/memphisSingleton"
	"github.com/memphisdev/memphis-rest-gateway/router"
	"github.com/nats-io/nats.go"
)

func Run(cfact sputnik.ConfFactory, lgr *logger.Logger, conn *nats.Conn) (stop func() error, err error) {

	var cnf conf.Configuration

	if err = cfact("connector", &cnf); err != nil {
		return nil, err
	}

	conf.Put(cnf)

	mconntr.Put(conn)

	err = handlers.ListenForUpdates(lgr)
	if err != nil {
		return nil, fmt.Errorf("Error while listening for updates - %s", err.Error())
	}

	go handlers.CleanConnectionsCache()
	app := router.SetupRoutes(lgr)
	lgr.Noticef("Memphis REST gateway is up and running as part of memphis-protocol-adapter")
	lgr.Noticef("Version %s", cnf.VERSION)

	go func() {
		app.Listen(":" + cnf.HTTP_PORT)
	}()

	return func() error {
		return app.ShutdownWithTimeout(time.Second * 5)
	}, nil
}
