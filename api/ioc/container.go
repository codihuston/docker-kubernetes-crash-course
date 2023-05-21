package ioc

import (
	logger "example.com/m/v2/pkg/logger"
)

type IOC struct {
	Logger *logger.DefaultLogger
}

// NewContainer returns a struct IOC (Inversion of Control) which contains
// concerns or dependencies that we'll use throughout the application via
// dependency injection. It might be possible that a future logger impl might
// depend on some storage configuration, which could be configured here in one
// place. We may consider adding our own abstraction layer between other new
// dependencies (like we did the logger) that might need to be available in the
// same manner. Another good example of things to put in here might be a
// global configuration of the application.
//
// Note: the database connection is not included in here as to keep its
// usage relevant to only the parts of the application that need it
// (repositories).
func NewContainer() IOC {
	c := IOC{}

	logger := logger.NewDefaultLogger()
	logger.Info("Welcome to my blogger app!")

	c.Logger = logger

	return c
}
