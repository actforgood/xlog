// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

import (
	"io"
)

// SyncLogger is a Logger which writes logs synchronously.
// It just calls underlying writer with each log call.
// Note: if used in a concurrent context, log writes are not concurrent safe,
// unless the writer is concurrent safe. See also NewSyncWriter on this matter.
type SyncLogger struct {
	// writer logs will be written to.
	writer io.Writer
	// formatter can be set with SyncLoggerWithFormatter functional option.
	formatter Formatter
	// common options for this logger.
	// can be set with SyncLoggerWithOptions functional option.
	opts *CommonOpts
}

// NewSyncLogger instantiates a new logger object that writes logs
// synchronously.
// First param is a Writer where logs are written to.
// Example: os.Stdout, a custom opened os.File, an in memory strings.Buffer, etc.
// Second param is/are function option(s) through which you can customize
// the logger. Check for SyncLoggerWith* options.
func NewSyncLogger(w io.Writer, opts ...SyncLoggerOption) *SyncLogger {
	// instantiate object with default properties.
	logger := &SyncLogger{
		writer:    w,
		formatter: JSONFormatter,
		opts:      NewCommonOpts(),
	}

	// apply functional options, if any.
	for _, opt := range opts {
		opt(logger)
	}

	return logger
}

// Critical logs application component unavailable, fatal events.
func (logger *SyncLogger) Critical(keyValues ...interface{}) {
	logger.log(LevelCritical, keyValues...)
}

// Error logs runtime errors that
// should typically be logged and monitored.
func (logger *SyncLogger) Error(keyValues ...interface{}) {
	logger.log(LevelError, keyValues...)
}

// Warn logs exceptional occurrences that are not errors.
// Example: Use of deprecated APIs, poor use of an API, undesirable things
// that are not necessarily wrong.
func (logger *SyncLogger) Warn(keyValues ...interface{}) {
	logger.log(LevelWarning, keyValues...)
}

// Info logs interesting events.
// Example: User logs in, SQL logs.
func (logger *SyncLogger) Info(keyValues ...interface{}) {
	logger.log(LevelInfo, keyValues...)
}

// Debug logs detailed debug information.
func (logger *SyncLogger) Debug(keyValues ...interface{}) {
	logger.log(LevelDebug, keyValues...)
}

// Log logs arbitrary data.
func (logger *SyncLogger) Log(keyValues ...interface{}) {
	logger.log(LevelNone, keyValues...)
}

// Close nicely closes logger.
func (logger *SyncLogger) Close() {
	if bw, ok := logger.writer.(*BufferedWriter); ok {
		bw.Stop()
	}
}

// log is used internally to write the log, if eligible.
// Default key-values are prepended to user passed ones.
func (logger *SyncLogger) log(lvl Level, keyValues ...interface{}) {
	// ignore log conditions check.
	if !logger.opts.BetweenMinMax(lvl) {
		return
	}

	// enrich passed key values with default ones.
	keyVals := logger.opts.WithDefaultKeyValues(lvl, keyValues...)

	// format the log.
	if err := logger.formatter(logger.writer, keyVals); err != nil {
		logger.opts.ErrHandler(err, keyVals)
	}
}
