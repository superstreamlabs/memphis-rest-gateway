package conf

import (
	"fmt"
)

type JWT_Config struct {
	JWT_SECRET                     string // The secret key used to generate the JWT token.
	JWT_EXPIRES_IN_MINUTES         int    // The JWT token valid time in minutes.
	REFRESH_JWT_SECRET             string // The secret key used to generate the JWT refresh token.
	REFRESH_JWT_EXPIRES_IN_MINUTES int    // The JWT refresh token valid time in minutes.
}

type API_Token_Config struct {
	API_TOKEN_HEADER string // Name of header that contains the API token.
	API_TOKEN        string // API token shared between the client and memphis-rest-gateway.

	HMAC_TOKEN_HEADER string // Name of header of the body signature. The body signature is a MAC hex digest of the body calcuated using the TOKEN_SECRET as the hash key.
	HMAC_TOKEN_SECRET string // A shared secret between the client and memphis-rest-gateway for calculating the body signature.
	HMAC_TOKEN_HASH   string // Hash algorithm for calculating the body signature, can be either sha256 or sha512.
}

type Stations_Config struct {
	NAME             string // Station name.
	AUTH_METHOD      string // Either "jwt", "api_token" , "hmac_token" or "none".
	JWT_Config              // JWT authentication settings.
	API_Token_Config        // API Token authentication settings.
}

type Configuration struct {
	VERSION              string            // Configuration version
	HTTP_PORT            string            // The port for the rest-gateway.
	MEMPHIS_HOST         string            // Memphis host eithe IP or FQDN.
	ROOT_CA_PATH         string            // Root CA of the Memphis server.
	CLIENT_CERT_PATH     string            // The rest-gateway certificate.
	CLIENT_KEY_PATH      string            // The rest-gateway privuate key.
	USER_PASS_BASED_AUTH bool              // If true, the initial connection to Memphis uses username and password for authentication, otherwise a connection token is used.
	ROOT_USER            string            // The username to use when USER_PASS_BASED_AUTH is true.
	ROOT_PASSWORD        string            // The password to use when USER_PASS_BASED_AUTH is true.
	CONNECTION_TOKEN     string            // The connection to use when USER_PASS_BASED_AUTH is false.
	DEV_ENV              string            // If "true", enables the /monitoring/getResourcesUtilization route.
	DEBUG                bool              // If true, enabled debug messages.
	AUTH_METHOD          string            // Global auth method, can be either "jwt", "api_token", "hmac_token" or "none".
	JWT_Config                             // Global jwt configuration if auth method is "jwt".
	API_Token_Config                       // Global api token configuration if auth method is either "api_token" or "hmac_token".
	Stations             []Stations_Config // Individual authentication method configuration for individual stations.

}

func GetConfig() Configuration {
	configuration := Configuration{}
	GonfigGetConf("./conf/config.json", &configuration)
	return configuration
}

func ValidateAuthMethod(cfg Configuration) error {
	switch cfg.AUTH_METHOD {
	case "jwt":
		if cfg.JWT_SECRET == "" {
			return fmt.Errorf("JWT_SECRET is either the empty string or missing")
		}
		if cfg.REFRESH_JWT_SECRET == "" {
			return fmt.Errorf("REFRESH_JWT_SECRET is either the empty string or missing")
		}

	case "api_token":
		if cfg.API_TOKEN == "" {
			return fmt.Errorf("API_TOKEN is either the empty string or missing")
		}
		if cfg.API_TOKEN_HEADER == "" {
			return fmt.Errorf("API_TOKEN_HEADER is either the empty string or missing")
		}

	case "hmac_token":
		if cfg.HMAC_TOKEN_SECRET == "" {
			return fmt.Errorf("HMAC_TOKEN_SECRET is either the empty string or missing")
		}
		if cfg.HMAC_TOKEN_HEADER == "" {
			return fmt.Errorf("HMAC_TOKEN_HEADER is either the empty string or missing")
		}
		if cfg.HMAC_TOKEN_HASH == "" {
			return fmt.Errorf("HMAC_TOKEN_HMAC_HASH is either the empty string or missing")
		}
	case "none":
		return nil
	default:
		if cfg.AUTH_METHOD == "" {
			return nil
		} else {
			return fmt.Errorf("AUTH_METHOD is '%v' but must be either 'jwt', 'api_token', 'hmac_token' or 'none'", cfg.AUTH_METHOD)
		}
	}
	return nil
}

func ValidateAuthMethodStation(cfg Stations_Config) error {
	switch cfg.AUTH_METHOD {
	case "jwt":
		if cfg.JWT_SECRET == "" {
			return fmt.Errorf("JWT_SECRET is either the empty string or missing")
		}
		if cfg.REFRESH_JWT_SECRET == "" {
			return fmt.Errorf("REFRESH_JWT_SECRET is either the empty string or missing")
		}

	case "api_token":
		if cfg.API_TOKEN == "" {
			return fmt.Errorf("API_TOKEN is either the empty string or missing")
		}
		if cfg.API_TOKEN_HEADER == "" {
			return fmt.Errorf("API_TOKEN_HEADER is either the empty string or missing")
		}

	case "hmac_token":
		if cfg.HMAC_TOKEN_SECRET == "" {
			return fmt.Errorf("HMAC_TOKEN_SECRET is either the empty string or missing")
		}
		if cfg.HMAC_TOKEN_HEADER == "" {
			return fmt.Errorf("HMAC_TOKEN_HEADER is either the empty string or missing")
		}
		if cfg.HMAC_TOKEN_HASH == "" {
			return fmt.Errorf("HMAC_TOKEN_HMAC_HASH is either the empty string or missing")
		}
	case "none":
		return nil
	default:
		return fmt.Errorf("AUTH_METHOD is '%v' but has to be either 'jwt', 'api_token', 'hmac_token' or 'none'", cfg.AUTH_METHOD)
	}
	return nil
}

func Validate(configuration Configuration) error {

	err := ValidateAuthMethod(configuration)
	if err != nil {
		return err
	}
	if configuration.Stations != nil {
		for i, station := range configuration.Stations {
			if station.NAME == "" {
				return fmt.Errorf("station '%v' is mising a name", i)
			}
			err = ValidateAuthMethodStation(station)
			if err != nil {
				return fmt.Errorf("station '%v' - %v", i, err)
			}
		}
	}

	return nil
}
