// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/actforgood/xlog"
)

func ExampleMultiLogger_splitMessagesByLevel() {
	// In this example we create a (multi)logger that writes
	// debug and info logs to standard output and
	// warning, error and critical logs to error output.

	errLgrOpts := xlog.NewCommonOpts()
	errLgrOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelWarning)
	errLgrOpts.Time = func() interface{} { // mock time for output check
		return "2022-03-20T16:01:20Z"
	}
	errLgrOpts.Source = xlog.SourceProvider(5, 1) // keep only filename for output check
	errLgr := xlog.NewSyncLogger(
		os.Stderr,
		xlog.SyncLoggerWithOptions(errLgrOpts),
	)

	dbgLgrOpts := xlog.NewCommonOpts()
	dbgLgrOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelDebug)
	dbgLgrOpts.MaxLevel = xlog.FixedLevelProvider(xlog.LevelInfo)
	dbgLgrOpts.Time = func() interface{} { // mock time for output check
		return "2022-03-20T16:01:20Z"
	}
	dbgLgrOpts.Source = xlog.SourceProvider(5, 1) // keep only filename for output check
	dbgLgr := xlog.NewSyncLogger(
		os.Stdout,
		xlog.SyncLoggerWithOptions(dbgLgrOpts),
	)

	logger := xlog.NewMultiLogger(errLgr, dbgLgr)
	defer logger.Close()

	logger.Debug("msg", "I get written to standard output")
	logger.Error("msg", "I get written to standard error")

	// Output:
	// {"date":"2022-03-20T16:01:20Z","lvl":"DEBUG","msg":"I get written to standard output","src":"/logger_multi_test.go:48"}
}

func ExampleMultiLogger_logToStdOutAndCustomFile() {
	// In this example we create a (multi)logger that writes
	// logs to standard output, and to a file.

	stdOutLgrOpts := xlog.NewCommonOpts()
	stdOutLgrOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	stdOutLgrOpts.Time = func() interface{} { // mock time for output check
		return "2022-03-15T16:01:20Z"
	}
	stdOutLgrOpts.Source = xlog.SourceProvider(5, 1) // keep only filename for output check
	stdOutLgr := xlog.NewSyncLogger(
		os.Stdout,
		xlog.SyncLoggerWithOptions(stdOutLgrOpts),
	)

	fileLgrOpts := xlog.NewCommonOpts()
	fileLgrOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	fileLgrOpts.Time = func() interface{} { // mock time for output check
		return "2022-03-15T16:01:20Z"
	}
	fileLgrOpts.Source = xlog.SourceProvider(5, 1) // keep only filename for output check
	f, err := ioutil.TempFile("", "x.log-")        // you will have your well known path defined
	if err != nil {
		panic(err)
	}
	fileLgr := xlog.NewSyncLogger(
		f,
		xlog.SyncLoggerWithOptions(fileLgrOpts),
	)

	logger := xlog.NewMultiLogger(stdOutLgr, fileLgr)
	defer func() {
		_ = logger.Close()
		_ = f.Close()
		_ = os.Remove(f.Name()) // you won't remove the file
	}()

	logger.Debug("msg", "I get written to standard output and to a file")

	// Output:
	// {"date":"2022-03-15T16:01:20Z","lvl":"DEBUG","msg":"I get written to standard output and to a file","src":"/logger_multi_test.go:92"}
}

func TestMultiLogger_logsOnEveryLogger(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		levels = []xlog.Level{
			xlog.LevelNone,
			xlog.LevelDebug,
			xlog.LevelInfo,
			xlog.LevelWarning,
			xlog.LevelError,
			xlog.LevelCritical,
		}
		loggers = []xlog.Logger{
			xlog.NewMockLogger(),
			xlog.NewMockLogger(),
			xlog.NewMockLogger(),
		}
		subject xlog.Logger = xlog.NewMultiLogger(loggers...)
		kv                  = getInputKeyValues()
	)
	defer func() {
		assertNil(t, subject.Close())
	}()

	for _, lvl := range levels {
		for _, logger := range loggers {
			lgr := logger.(*xlog.MockLogger)
			lgr.SetLogCallback(lvl, func(keyValues ...interface{}) {
				assertEqual(t, kv, keyValues)
			})
		}

		// act
		callMethodByLevel(subject, lvl)

		// assert
		for _, logger := range loggers {
			lgr := logger.(*xlog.MockLogger)
			assertEqual(t, 1, lgr.LogCallsCount(lvl))
		}
	}
}

func TestMultiLogger_Close_closesAllLoggers(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		logger1            = xlog.NewMockLogger()
		logger2            = xlog.NewMockLogger()
		logger3            = xlog.NewMockLogger()
		loggers            = []xlog.Logger{logger1, logger2, logger3}
		subject            = xlog.NewMultiLogger(loggers...)
		expectedLogger1Err = errors.New("intentionally triggered logger 1 Close error")
		expectedLogger3Err = errors.New("intentionally triggered logger 3 Close error")
	)
	logger1.SetCloseError(expectedLogger1Err)
	logger3.SetCloseError(expectedLogger3Err)

	// act
	err := subject.Close()

	// assert
	if assertNotNil(t, err) {
		assertTrue(t, errors.Is(err, expectedLogger1Err))
		assertTrue(t, errors.Is(err, expectedLogger3Err))
	}
	for _, logger := range loggers {
		lgr := logger.(*xlog.MockLogger)
		assertEqual(t, 1, lgr.CloseCallsCount())
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelNone))
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelDebug))
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelInfo))
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelWarning))
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelError))
		assertEqual(t, 0, lgr.LogCallsCount(xlog.LevelCritical))
	}
}
