package conf

import (
	"fmt"

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
	CLIENT_CERT_PATH               string
	CLIENT_KEY_PATH                string
	ROOT_CA_PATH                   string
	USER_PASS_BASED_AUTH           bool
	ROOT_PASSWORD                  string
	DEBUG                          bool
	AUTH_METHOD                    string // Either "jwt", "api_token" , "hmac_token" or "none".
	API_TOKEN                      string // API token shared between the client and memphis-rest-gateway.
	API_TOKEN_HEADER               string // Name of header that contains the API token.
	HMAC_TOKEN_SECRET              string // A shared secret between the client and memphis-rest-gateway.
	HMAC_TOKEN_HEADER              string // Name of header that contains the body signature which is a MAC hex digest of the body calcuated using the TOKEN_SECRET as the hash key.
	HMAC_TOKEN_HASH                string // Hash algorithm for cacluating body signature, can be either sha256 or sha512.

}

func GetConfig() Configuration {
	configuration := Configuration{}
	gonfig.GetConf("./conf/config.json", &configuration)

	return configuration
}

func Validate(configuration Configuration) error {

	switch configuration.AUTH_METHOD {
	case "jwt":
		// JWT_EXPIRES_IN_MINUTES and REFRESH_JWT_EXPIRES_IN_MINUTES defaults to 0 if they are missing from configuration file.
		// Don't know if that should generate an error or not.

		if configuration.JWT_SECRET == "" {
			return fmt.Errorf("configuration option JWT_SECRET is either the empty string or missing")
		}
		if configuration.REFRESH_JWT_SECRET == "" {
			return fmt.Errorf("configuration option REFRESH_JWT_SECRET is either the empty string or missing")
		}

	case "api_token":
		if configuration.API_TOKEN == "" {
			return fmt.Errorf("configuration option API_TOKEN is either the empty string or missing")
		}
		if configuration.API_TOKEN_HEADER == "" {
			return fmt.Errorf("configuration option API_TOKEN_HEADER is either the empty string or missing")
		}

	case "hmac_token":
		if configuration.HMAC_TOKEN_SECRET == "" {
			return fmt.Errorf("configuration option HMAC_TOKEN_SECRET is either the empty string or missing")
		}
		if configuration.HMAC_TOKEN_HEADER == "" {
			return fmt.Errorf("configuration option HMAC_TOKEN_HEADER is either the empty string or missing")
		}
		if configuration.HMAC_TOKEN_HASH == "" {
			return fmt.Errorf("configuration option HMAC_TOKEN_HMAC_HASH is either the empty string or missing")
		}
	case "none":
		return nil
	default:
		if configuration.AUTH_METHOD == "" {
			return nil
		} else {
			return fmt.Errorf("configuration option AUTH_METHOD has to be either 'jwt', 'api_token', 'hmac_token' or 'none'")
		}
	}

	return nil
}
