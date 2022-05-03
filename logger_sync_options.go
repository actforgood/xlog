// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

// SyncLoggerOption defines optional function for configuring
// a sync logger.
type SyncLoggerOption func(*SyncLogger)

// SyncLoggerWithFormatter sets desired formatter.
// The JSON formatter is used by default.
func SyncLoggerWithFormatter(formatter Formatter) SyncLoggerOption {
	return func(logger *SyncLogger) {
		logger.formatter = formatter
	}
}

// SyncLoggerWithOptions sets the common options.
// A NewCommonOpts is used by default.
func SyncLoggerWithOptions(opts *CommonOpts) SyncLoggerOption {
	return func(logger *SyncLogger) {
		logger.opts = opts
	}
}
