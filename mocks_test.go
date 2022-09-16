// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/actforgood/xlog"
)

// Note: this file contains (internal) mocks needed in UTs.

// ErrWrite is a predefined error returned by WriteCallbackErr.
var ErrWrite = errors.New("intentionally triggered Writer error")

// Writer is a mock for io.Writer contract.
type MockWriter struct {
	writeCallsCnt uint32
	writeCallback func(p []byte) (n int, err error)
}

// Write mock logic.
func (mock *MockWriter) Write(p []byte) (int, error) {
	atomic.AddUint32(&mock.writeCallsCnt, 1)
	if mock.writeCallback != nil {
		return mock.writeCallback(p)
	}

	return len(p), nil
}

// SetWriteCallback sets the callback to be executed inside Write.
// You can hook to control the returned value(s), to make assertions
// upon passed parameter(s).
func (mock *MockWriter) SetWriteCallback(
	callback func(p []byte) (n int, err error),
) {
	mock.writeCallback = callback
}

// WriteCallsCount returns the no. of times Write was called.
func (mock *MockWriter) WriteCallsCount() int {
	return int(atomic.LoadUint32(&mock.writeCallsCnt))
}

// WriteCallbackErr is a predefined Write callback that returns
// an error.
func WriteCallbackErr(_ []byte) (int, error) {
	return 0, ErrWrite
}

// ErrFormat is a predefined error returned by FormatCallbackErr.
var ErrFormat = errors.New("intentionally triggered Formatter error")

// Formatter is a mocked wrapper for a xlog.Formatter.
type MockFormatter struct {
	formatCallsCnt uint32
	formatCallback xlog.Formatter
}

// Format is the Formatter function.
func (mock *MockFormatter) Format(w io.Writer, keyValues []interface{}) error {
	atomic.AddUint32(&mock.formatCallsCnt, 1)
	if mock.formatCallback != nil {
		return mock.formatCallback(w, keyValues)
	}

	return nil
}

// SetFormatCallback sets the callback to be executed inside Format.
// You can hook to control the returned value(s), to make assertions
// upon passed parameters.
func (mock *MockFormatter) SetFormatCallback(callback xlog.Formatter) {
	mock.formatCallback = callback
}

// FormatCallsCount returns the no. of times Format was called.
func (mock *MockFormatter) FormatCallsCount() int {
	return int(atomic.LoadUint32(&mock.formatCallsCnt))
}

// FormatCallbackErr is a predefined MockFormatter callback that returns
// an error.
func FormatCallbackErr(_ io.Writer, _ []interface{}) error {
	return ErrFormat
}

// MockErrorHandler is a mocked wrapper for an xlog.ErrorHandler.
type MockErrorHandler struct {
	handleCallsCnt uint32
	handleCallback xlog.ErrorHandler
}

// Handle is the ErrorHandler function.
func (mock *MockErrorHandler) Handle(err error, keyValues []interface{}) {
	atomic.AddUint32(&mock.handleCallsCnt, 1)
	if mock.handleCallback != nil {
		mock.handleCallback(err, keyValues)
	}
}

// SetHandleCallback sets the callback to be executed inside Handle.
// You can hook into to make assertions upon passed parameters.
func (mock *MockErrorHandler) SetHandleCallback(callback xlog.ErrorHandler) {
	mock.handleCallback = callback
}

// HandleCallsCount returns the no. of times Handle was called.
func (mock *MockErrorHandler) HandleCallsCount() int {
	return int(atomic.LoadUint32(&mock.handleCallsCnt))
}
