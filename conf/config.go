package conf

import (
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	VERSION                        string
	JWT_SECRET                     string
	JWT_EXPIRES_IN_MINUTES         int
	REFRESH_JWT_SECRET             string
	REFRESH_JWT_EXPIRES_IN_MINUTES int
	HTTP_PORT                      string
	ROOT_USER                      string
	CONNECTION_TOKEN               string
	MEMPHIS_HOST                   string
	DEV_ENV                        string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	gonfig.GetConf("./conf/config.json", &configuration)

	return configuration
}
