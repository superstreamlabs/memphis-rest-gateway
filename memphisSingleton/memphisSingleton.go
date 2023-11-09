package memphisSingleton

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/memphisdev/memphis-rest-gateway/conf"

	"github.com/nats-io/nats.go"
)

var nconn atomic.Pointer[nats.Conn]

func Get() (*nats.Conn, error) {
	if conn := nconn.Load(); conn != nil {
		return conn, nil
	}

	return nil, fmt.Errorf("not connected to memphis")
}

func Put(conn *nats.Conn) {
	nconn.Store(conn)
}

func Connect(configuration conf.Configuration) (*nats.Conn, error) {
	creds := configuration.CONNECTION_TOKEN
	username := configuration.ROOT_USER
	if configuration.USER_PASS_BASED_AUTH {
		username = "$$memphis"
		creds = configuration.CONNECTION_TOKEN + "_" + configuration.ROOT_PASSWORD
		if !configuration.CLOUD_ENV {
			creds = configuration.ROOT_PASSWORD
		}
	}

	return connect(configuration.MEMPHIS_HOST, username, creds, &configuration)
}

func connect(hostname, username, creds string, configuration *conf.Configuration) (*nats.Conn, error) {
	var nc *nats.Conn
	var err error

	natsOpts := nats.Options{
		Url:            hostname + ":" + strconv.Itoa(configuration.MEMPHIS_PORT),
		AllowReconnect: true,
		MaxReconnect:   10,
		ReconnectWait:  3 * time.Second,
		Name:           configuration.MEMPHIS_CLIENT,
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

	Put(nc)

	return nc, nil
}
