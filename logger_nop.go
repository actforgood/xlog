// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

// NopLogger is a no-operation Logger which does nothing.
// It simply ignores any log.
type NopLogger struct{}

// Critical logs application component unavailable, fatal events.
func (NopLogger) Critical(...interface{}) {}

// Error logs runtime errors that
// should typically be logged and monitored.
func (NopLogger) Error(...interface{}) {}

// Warn logs exceptional occurrences that are not errors.
// Example: Use of deprecated APIs, poor use of an API, undesirable things
// that are not necessarily wrong.
func (NopLogger) Warn(...interface{}) {}

// Info logs interesting events.
// Example: User logs in, SQL logs.
func (NopLogger) Info(...interface{}) {}

// Debug logs detailed debug information.
func (NopLogger) Debug(...interface{}) {}

// Log logs arbitrary data.
func (NopLogger) Log(...interface{}) {}

// Close nicely closes logger.
func (NopLogger) Close() error { return nil }
