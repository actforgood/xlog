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
	logCallsCnt     map[Level]uint32
	logCallbacks    map[Level]func(keyValues ...any)
	loggedKeyValues [][]any
	closeCallsCnt   uint32
	closeErr        error
	mu              sync.RWMutex
}

// MockKeyNotFound is the type of value returned by [MockLogger.ValueAt]
// in case the searched key is not found.
type MockKeyNotFound struct{}

// MockAny can be passed to [MockLogger.ValueAt] as call index,
// in case the order is not known/important/multiple logs happen concurrently.
const MockAny uint = 0

// NewMockLogger instantiates new mocked Logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logCallsCnt:  make(map[Level]uint32, 5),
		logCallbacks: make(map[Level]func(keyValues ...any), 5),
	}
}

// Critical mock logic.
func (mock *MockLogger) Critical(keyValues ...any) {
	mock.logByLevel(LevelCritical, keyValues...)
}

// Error mock logic.
func (mock *MockLogger) Error(keyValues ...any) {
	mock.logByLevel(LevelError, keyValues...)
}

// Warn mock logic.
func (mock *MockLogger) Warn(keyValues ...any) {
	mock.logByLevel(LevelWarning, keyValues...)
}

// Info mock logic.
func (mock *MockLogger) Info(keyValues ...any) {
	mock.logByLevel(LevelInfo, keyValues...)
}

// Debug mock logic.
func (mock *MockLogger) Debug(keyValues ...any) {
	mock.logByLevel(LevelDebug, keyValues...)
}

// Log mock logic.
func (mock *MockLogger) Log(keyValues ...any) {
	mock.logByLevel(LevelNone, keyValues...)
}

func (mock *MockLogger) logByLevel(lvl Level, keyValues ...any) {
	mock.mu.Lock()
	mock.logCallsCnt[lvl]++
	mock.mu.Unlock()

	if mock.logCallbacks[lvl] != nil {
		mock.logCallbacks[lvl](keyValues...)
	} else {
		// default behaviour is to store values in a slice.
		// values can be retrieved later with ValueAt.
		mock.mu.Lock()
		mock.loggedKeyValues = append(mock.loggedKeyValues, keyValues)
		mock.mu.Unlock()
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
	callback func(keyValues ...any),
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

// ValueAt returns the value for a key at given call, in case no callback was set.
// Calls are positive numbers (starting with 1).
// If the order of the calls is not known/important/multiple logs happen concurrently,
// you can use [MockAtAnyCall].
// If the key is not found,  [MockKeyNotFound] is returned.
func (mock *MockLogger) ValueAt(callNo uint, forKey any) any {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	if callNo == MockAny {
		for call := range len(mock.loggedKeyValues) {
			value := mock.valueAt(call+1, forKey)
			if _, isNotFound := value.(MockKeyNotFound); !isNotFound {
				return value
			}
		}

		return MockKeyNotFound{}
	}

	return mock.valueAt(int(callNo), forKey)
}

func (mock *MockLogger) valueAt(callNo int, forKey any) any {
	if len(mock.loggedKeyValues) >= callNo {
		for i := 0; i <= len(mock.loggedKeyValues[callNo-1])-2; i += 2 {
			if mock.loggedKeyValues[callNo-1][i] == forKey && len(mock.loggedKeyValues[callNo-1]) > i {
				return mock.loggedKeyValues[callNo-1][i+1]
			}
		}
	}

	return MockKeyNotFound{}
}
