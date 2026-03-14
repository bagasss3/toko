package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(service string) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	// attach service name to every log entry
	log.WithField("service", service)

	return log
}

// NewEntry returns a logrus.Entry with service field already attached
func NewEntry(service string) *logrus.Entry {
	return New(service).WithField("service", service)
}
