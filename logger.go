// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

// Logger provides prototype for logging with different levels.
// It is designed to accept variadic parameters useful for a
// structured logger.
type Logger interface {
	// Critical logs application component unavailable, fatal events.
	Critical(keyValues ...interface{})

	// Error logs runtime errors that
	// should typically be logged and monitored.
	Error(keyValues ...interface{})

	// Warn logs exceptional occurrences that are not errors.
	// Example: Use of deprecated APIs, poor use of an API,
	// undesirable things that are not necessarily wrong.
	Warn(keyValues ...interface{})

	// Info logs interesting events.
	// Example: User logs in, SQL logs.
	Info(keyValues ...interface{})

	// Debug logs detailed debug information.
	Debug(keyValues ...interface{})

	// Log logs arbitrary data.
	Log(keyValues ...interface{})

	// Close performs clean up actions, closes resources,
	// avoids memory leaks, etc.
	// Make sure to call it at your application shutdown
	// for example.
	Close()
}
