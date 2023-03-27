package logger

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"rest-gateway/conf"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
)

const (
	restGwSourceName   = "rest-gateway"
	syslogsStreamName  = "$memphis_syslogs"
	syslogsInfoSubject = "extern.info"
	syslogsWarnSubject = "extern.warn"
	syslogsErrSubject  = "extern.err"
	labelLen           = 3
	infoLabel          = "[INF] "
	debugLabel         = "[DBG] "
	warnLabel          = "[WRN] "
	errorLabel         = "[ERR] "
	fatalLabel         = "[FTL] "
	traceLabel         = "[TRC] "
)

type streamWriter struct {
	nc         *nats.Conn
	labelStart int
}

type Logger struct {
	logger *log.Logger
}

func (sw streamWriter) Write(p []byte) (int, error) {
	os.Stderr.Write(p)
	logLabelToSubjectMap := map[string]string{"INF": syslogsInfoSubject,
		"WRN": syslogsWarnSubject,
		"ERR": syslogsErrSubject}

	label := string(p[sw.labelStart : sw.labelStart+labelLen])
	subjectSuffix, ok := logLabelToSubjectMap[label]
	if !ok { // skip other labels
		return 0, nil
	}

	subject := fmt.Sprintf("%s.%s.%s", syslogsStreamName, restGwSourceName, subjectSuffix)
	err := sw.nc.Publish(subject, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func CreateLogger(hostname string, username string, creds string) (*Logger, error) {
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

		pemData, err := ioutil.ReadFile(configuration.ROOT_CA_PATH)
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

	flags := log.LstdFlags | log.Lmicroseconds
	pidPrefix := fmt.Sprintf("[%d] ", os.Getpid())
	labelStart := len(pidPrefix) + 28

	sw := streamWriter{
		nc:         nc,
		labelStart: labelStart,
	}

	return &Logger{
		logger: log.New(sw, pidPrefix, flags),
	}, nil
}

func SetLogger(app *fiber.App, l *Logger) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", l)
		return c.Next()
	})
}

func GetLogger(c *fiber.Ctx) *Logger {
	return c.Locals("logger").(*Logger)
}

// Noticef logs a notice statement
func (l *Logger) Noticef(format string, v ...interface{}) {
	l.logger.Printf(infoLabel+format, v...)
}

// Warnf logs a warning statement
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Printf(warnLabel+format, v...)
}

// Errorf logs an error statement
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Printf(errorLabel+format, v...)
}

// Fatalf logs a fatal error
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(fatalLabel+format, v...)
}

// Debugf logs a debug statement
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Printf(debugLabel+format, v...)

}

// Tracef logs a trace statement
func (l *Logger) Tracef(format string, v ...interface{}) {
	l.logger.Printf(traceLabel+format, v...)
}
