// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import "github.com/actforgood/xerr"

// MultiLogger is a composite Logger capable of
// logging to multiple loggers.
type MultiLogger struct {
	// loggers to log messages to.
	loggers []Logger
}

// NewMultiLogger instantiates a new multi logger object.
// Accepts the loggers multi-logger handles.
func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// Critical logs application component unavailable, fatal events.
func (logger *MultiLogger) Critical(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Critical(keyValues...)
	}
}

// Error logs runtime errors that
// should typically be logged and monitored.
func (logger *MultiLogger) Error(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Error(keyValues...)
	}
}

// Warn logs exceptional occurrences that are not errors.
// Example: Use of deprecated APIs, poor use of an API, undesirable things
// that are not necessarily wrong.
func (logger *MultiLogger) Warn(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Warn(keyValues...)
	}
}

// Info logs interesting events.
// Example: User logs in, SQL logs.
func (logger *MultiLogger) Info(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Info(keyValues...)
	}
}

// Debug logs detailed debug information.
func (logger *MultiLogger) Debug(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Debug(keyValues...)
	}
}

// Log logs arbitrarily data.
func (logger *MultiLogger) Log(keyValues ...interface{}) {
	for _, lgr := range logger.loggers {
		lgr.Log(keyValues...)
	}
}

// Close performs clean up actions, closes resources,
// avoids memory leaks, etc.
// Make sure to call it at your application shutdown
// for example.
func (logger *MultiLogger) Close() error {
	var mErr *xerr.MultiError
	for _, lgr := range logger.loggers {
		if err := lgr.Close(); err != nil {
			mErr = mErr.Add(err)
		}
	}

	return mErr.ErrOrNil()
}
