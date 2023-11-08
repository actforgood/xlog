// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

// NopLogger is a no-operation Logger which does nothing.
// It simply ignores any log.
type NopLogger struct{}

// Critical logs application component unavailable, fatal events.
func (NopLogger) Critical(...any) {}

// Error logs runtime errors that
// should typically be logged and monitored.
func (NopLogger) Error(...any) {}

// Warn logs exceptional occurrences that are not errors.
// Example: Use of deprecated APIs, poor use of an API, undesirable things
// that are not necessarily wrong.
func (NopLogger) Warn(...any) {}

// Info logs interesting events.
// Example: User logs in, SQL logs.
func (NopLogger) Info(...any) {}

// Debug logs detailed debug information.
func (NopLogger) Debug(...any) {}

// Log logs arbitrary data.
func (NopLogger) Log(...any) {}

// Close nicely closes logger.
func (NopLogger) Close() error { return nil }
