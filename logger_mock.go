// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"sync"
)

// MockLogger is a mock for xlog.Logger contract, to be used in UT.
type MockLogger struct {
	logCallsCnt   map[Level]uint32
	logCallbacks  map[Level]func(keyValues ...interface{})
	closeCallsCnt uint32
	closeErr      error
	mu            sync.RWMutex
}

// NewMockLogger instantiates new mocked Logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logCallsCnt:  make(map[Level]uint32, 5),
		logCallbacks: make(map[Level]func(keyValues ...interface{}), 5),
	}
}

// Critical mock logic.
func (mock *MockLogger) Critical(keyValues ...interface{}) {
	mock.logByLevel(LevelCritical, keyValues...)
}

// Error mock logic.
func (mock *MockLogger) Error(keyValues ...interface{}) {
	mock.logByLevel(LevelError, keyValues...)
}

// Warn mock logic.
func (mock *MockLogger) Warn(keyValues ...interface{}) {
	mock.logByLevel(LevelWarning, keyValues...)
}

// Info mock logic.
func (mock *MockLogger) Info(keyValues ...interface{}) {
	mock.logByLevel(LevelInfo, keyValues...)
}

// Debug mock logic.
func (mock *MockLogger) Debug(keyValues ...interface{}) {
	mock.logByLevel(LevelDebug, keyValues...)
}

// Log mock logic.
func (mock *MockLogger) Log(keyValues ...interface{}) {
	mock.logByLevel(LevelNone, keyValues...)
}

func (mock *MockLogger) logByLevel(lvl Level, keyValues ...interface{}) {
	mock.mu.Lock()
	mock.logCallsCnt[lvl]++
	mock.mu.Unlock()

	if mock.logCallbacks[lvl] != nil {
		mock.logCallbacks[lvl](keyValues...)
	}
}

// Close mock logic.
func (mock *MockLogger) Close() error {
	mock.mu.Lock()
	mock.closeCallsCnt++
	mock.mu.Unlock()

	return mock.closeErr
}

// SetLogCallback sets the callback to be executed inside Error/Warn/Info/Debug/Log.
// You can make assertions upon passed parameter(s) this way.
func (mock *MockLogger) SetLogCallback(
	lvl Level,
	callback func(keyValues ...interface{}),
) {
	mock.logCallbacks[lvl] = callback
}

// SetCloseError sets the error to be returned by the Close method.
func (mock *MockLogger) SetCloseError(closeErr error) {
	mock.closeErr = closeErr
}

// LogCallsCount returns the no. of times Critical/Error/Warn/Info/Debug/Log was called.
// Differentiate methods calls count by passing appropriate level.
func (mock *MockLogger) LogCallsCount(lvl Level) int {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	return int(mock.logCallsCnt[lvl])
}

// CloseCallsCount returns the no. of times Close was called.
func (mock *MockLogger) CloseCallsCount() int {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	return int(mock.closeCallsCnt)
}
