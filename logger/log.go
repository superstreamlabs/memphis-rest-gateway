package logger

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/dapr/kit/logger"
	"github.com/memphisdev/memphis-rest-gateway/conf"
	"github.com/memphisdev/memphis-rest-gateway/memphisSingleton"

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
	configuration := conf.Get()
	if !configuration.CLOUD_ENV {
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
	}
	return len(p), nil
}

func CreateLogger(hostname string, username string, creds string) (*Logger, error) {
	mc, err := memphisSingleton.Get()
	if err != nil {
		return nil, err
	}

	flags := log.LstdFlags | log.Lmicroseconds
	pidPrefix := fmt.Sprintf("[%d] ", os.Getpid())
	labelStart := len(pidPrefix) + 28

	sw := streamWriter{
		nc:         mc,
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

var lgr atomic.Pointer[logger.Logger]

func Get() *logger.Logger {
	return lgr.Load()
}

func Put(logger *logger.Logger) {
	lgr.Store(logger)
}
