package conf

type Configuration struct {
	VERSION                        string
	MEMPHIS_HOST                   string
	MEMPHIS_PORT                   int
	MEMPHIS_CLIENT                 string
	USER_PASS_BASED_AUTH           bool
	ROOT_PASSWORD                  string
	ROOT_USER                      string
	CONNECTION_TOKEN               string
	CLIENT_CERT_PATH               string
	CLIENT_KEY_PATH                string
	ROOT_CA_PATH                   string
	JWT_EXPIRES_IN_MINUTES         int
	REFRESH_JWT_EXPIRES_IN_MINUTES int
	REST_GW_UPDATES_SUBJ           string
	JWT_SECRET                     string
	REFRESH_JWT_SECRET             string
	HTTP_PORT                      string
	DEV_ENV                        string
	DEBUG                          bool
	CLOUD_ENV                      bool
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
