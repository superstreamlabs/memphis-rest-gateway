package memphisSingleton

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"rest-gateway/conf"
	"time"

	"github.com/nats-io/nats.go"
)

var mc *nats.Conn

func GetMemphisConnection(hostname, creds, username string) (*nats.Conn, error) {
	if mc == nil {
		configuration := conf.GetConfig()
		var nc *nats.Conn
		var err error

		natsOpts := nats.Options{
			Url:            hostname + ":6666",
			AllowReconnect: true,
			MaxReconnect:   10,
			ReconnectWait:  3 * time.Second,
			Name:           "MEMPHIS HTTP LOGGER",
		}

		if configuration.USER_PASS_BASED_AUTH {
			natsOpts.Password = creds
			natsOpts.User = username
		} else {
			natsOpts.Token = username + "::" + creds
		}

		if configuration.CLIENT_CERT_PATH != "" && configuration.CLIENT_KEY_PATH != "" && configuration.ROOT_CA_PATH != "" {
			cert, err := tls.LoadX509KeyPair(configuration.CLIENT_CERT_PATH, configuration.CLIENT_KEY_PATH)
			if err != nil {
				return nil, err
			}
			cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
			if err != nil {
				return nil, err
			}
			TLSConfig := &tls.Config{MinVersion: tls.VersionTLS12}
			TLSConfig.Certificates = []tls.Certificate{cert}
			certs := x509.NewCertPool()

			pemData, err := os.ReadFile(configuration.ROOT_CA_PATH)
			if err != nil {
				return nil, err
			}
			certs.AppendCertsFromPEM(pemData)
			TLSConfig.RootCAs = certs
			natsOpts.TLSConfig = TLSConfig
		}

		nc, err = natsOpts.Connect()
		if err != nil {
			return nil, err
		}

		mc = nc
	}

	return mc, nil
}
