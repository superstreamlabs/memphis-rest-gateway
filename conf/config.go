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
	MEMPHIS_PORT                   int
	DEV_ENV                        string
	CLIENT_CERT_PATH               string
	CLIENT_KEY_PATH                string
	ROOT_CA_PATH                   string
	USER_PASS_BASED_AUTH           bool
	ROOT_PASSWORD                  string
	DEBUG                          bool
	CLOUD_ENV                      bool
	REST_GW_UPDATES_SUBJ           string
	MEMPHIS_CLIENT                 string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	gonfig.GetConf("./conf/config.json", &configuration)

	return configuration
}

var config = Configuration{}

func Get() Configuration {
	return config
}

func Put(cnf Configuration) {
	config = cnf
}

func Access() *Configuration {
	return &config
}
