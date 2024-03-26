//go:build !windows && !nacl && !plan9
// +build !windows,!nacl,!plan9

// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/actforgood/xlog"
)

func ExampleSyncLogger_withSyslog() {
	// In this example we create a SyncLogger that logs to syslog.

	syslogWriter, err := syslog.Dial("", "", syslog.LOG_ERR, "demo")
	if err != nil {
		panic(err)
	}

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelDebug)
	logger := xlog.NewSyncLogger(
		syslogWriter,
		xlog.SyncLoggerWithFormatter(xlog.SyslogFormatter(
			xlog.JSONFormatter,
			xlog.NewDefaultSyslogLevelProvider(opts),
			"",
		)),
		xlog.SyncLoggerWithOptions(opts),
	)
	defer func() {
		_ = logger.Close()
		_ = syslogWriter.Close()
	}()

	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)

	// sudo tail -f /var/log/syslog
	// Apr  1 03:03:16 bogdan-Aspire-5755G demo[7572]: {"date":"2022-04-01T00:03:16.209891806Z","lvl":"INFO","msg":"Hello World","src":"/home/bogdan/work/go/xlog/formatter_decorator_syslog_test.go:45","year":2022}
}

func ExampleSyncLogger_withSyslogSupportingAllSyslogLevels() {
	// In this example we create a SyncLogger that logs to syslog,
	// supporting all syslog levels. We can log extra syslog
	// levels with Log() method. This is just an example, based on "lvl"
	// key and Log() method.

	syslogWriter, err := syslog.Dial("", "", syslog.LOG_ERR, "demo")
	if err != nil {
		panic(err)
	}

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone) // need LevelNone to be able to use Log().
	allSyslogLevelsMap := map[any]syslog.Priority{          // we define all levels map.
		// default xlog levels/labels
		"CRITICAL": syslog.LOG_CRIT,
		"ERROR":    syslog.LOG_ERR,
		"WARN":     syslog.LOG_WARNING,
		"INFO":     syslog.LOG_INFO,
		"DEBUG":    syslog.LOG_DEBUG,
		// define extra syslog levels which will be logged with Log()
		"EMERG":  syslog.LOG_EMERG,
		"ALERT":  syslog.LOG_EMERG,
		"NOTICE": syslog.LOG_NOTICE,
	}
	logger := xlog.NewSyncLogger(
		syslogWriter,
		xlog.SyncLoggerWithFormatter(xlog.SyslogFormatter(
			xlog.JSONFormatter,
			xlog.NewExtractFromKeySyslogLevelProvider( // we set syslog level provider
				opts.LevelKey,
				allSyslogLevelsMap,
			),
			"",
		)),
		xlog.SyncLoggerWithOptions(opts),
	)
	defer func() {
		_ = logger.Close()
		_ = syslogWriter.Close()
	}()

	// log xlog levels through dedicated APIs as usual.
	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)
	// log extra syslog levels through Log().
	logger.Log("lvl", "NOTICE", xlog.MessageKey, "Hello World", "year", 2022)
}

func TestSyslogFormatter_successfullyWritesKeyValuesByLevel(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		commOpts    = xlog.NewCommonOpts()
		formatter   = new(MockFormatter)
		writer      = NewMockSyslogWriter()
		levelLabels = map[any]syslog.Priority{
			"DEBUG":    syslog.LOG_DEBUG,
			"INFO":     syslog.LOG_INFO,
			"NOTICE":   syslog.LOG_NOTICE,
			"WARN":     syslog.LOG_WARNING,
			"ERROR":    syslog.LOG_ERR,
			"CRITICAL": syslog.LOG_CRIT,
			"ALERT":    syslog.LOG_ALERT,
			"EMERG":    syslog.LOG_EMERG,
		}
		allLevelsProvider = xlog.NewExtractFromKeySyslogLevelProvider(commOpts.LevelKey, levelLabels)

		subject = xlog.SyslogFormatter(
			formatter.Format,
			allLevelsProvider,
			xlog.SyslogPrefixCee,
		)
	)

	idx := 0
	for levelLabel, syslogLvl := range levelLabels {
		testLvl := syslogLvl // capture range variable
		testLabel := levelLabel.(string)
		idx++
		t.Run(fmt.Sprintf("level_%s", testLabel), func(t *testing.T) {
			keyValues := []any{"abc", "123", commOpts.LevelKey, testLabel}
			formatter.SetFormatCallback(func(w io.Writer, kv []any) error {
				assertEqual(t, keyValues, kv)
				buf, ok := w.(*bytes.Buffer)
				if assertTrue(t, ok) {
					_, _ = buf.Write([]byte("some_kv_formatted_" + testLabel))
				}

				return nil
			})
			writer.SetLogCallback(testLvl, func(msg string) error {
				assertEqual(t, "@cee:some_kv_formatted_"+testLabel, msg)

				return nil
			})

			// act
			resultErr := subject(writer, keyValues)

			// assert
			assertNil(t, resultErr)
			assertEqual(t, idx, formatter.FormatCallsCount())
			assertEqual(t, 1, writer.LogCallsCount(testLvl))
		})
	}
}

func TestSyslogFormatter_successfullyWritesKeyValuesNoLevel(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		commOpts  = xlog.NewCommonOpts()
		formatter = new(MockFormatter)
		writer    = NewMockSyslogWriter()
		subject   = xlog.SyslogFormatter(
			formatter.Format,
			xlog.NewDefaultSyslogLevelProvider(commOpts),
			"",
		)
		keyValues = []any{"abc", "123", "no", "level"}
	)
	formatter.SetFormatCallback(func(w io.Writer, kv []any) error {
		assertEqual(t, keyValues, kv)
		buf, ok := w.(*bytes.Buffer)
		if assertTrue(t, ok) {
			_, _ = buf.Write([]byte("some_kv_formatted"))
		}

		return nil
	})
	writer.SetWriteCallback(func(p []byte) (int, error) {
		assertEqual(t, []byte("some_kv_formatted"), p)

		return len(p), nil
	})

	// act
	resultErr := subject(writer, keyValues)

	// assert
	assertNil(t, resultErr)
	assertEqual(t, 1, formatter.FormatCallsCount())
	assertEqual(t, 1, writer.WriteCallsCount())
}

func TestSyslogFormatter_returnsErrFromFormatter(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		commOpts  = xlog.NewCommonOpts()
		formatter = new(MockFormatter)
		writer    = NewMockSyslogWriter()
		subject   = xlog.SyslogFormatter(
			formatter.Format,
			xlog.NewDefaultSyslogLevelProvider(commOpts),
			"",
		)
		keyValues = []any{"foo", "bar", "year", 2022}
	)
	formatter.SetFormatCallback(FormatCallbackErr)

	// act
	resultErr := subject(writer, keyValues)

	// assert
	assertTrue(t, errors.Is(resultErr, ErrFormat))
	assertEqual(t, 1, formatter.FormatCallsCount())
	assertEqual(t, 0, writer.WriteCallsCount())
}

func TestSyslogFormatter_returnsErrNotSyslogWriter(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		commOpts  = xlog.NewCommonOpts()
		formatter = new(MockFormatter)
		writer    = io.Discard
		subject   = xlog.SyslogFormatter(
			formatter.Format,
			xlog.NewDefaultSyslogLevelProvider(commOpts),
			"",
		)
		keyValues = []any{"foo", "bar", "year", 2022}
	)

	// act
	resultErr := subject(writer, keyValues)

	// assert
	assertTrue(t, errors.Is(resultErr, xlog.ErrNotSyslogWriter))
	assertEqual(t, 0, formatter.FormatCallsCount())
}

func TestSyslogFormatter_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		commOpts = xlog.NewCommonOpts()
		writer   = NewMockSyslogWriter()
		subject  = xlog.SyslogFormatter(
			xlog.JSONFormatter,
			xlog.NewDefaultSyslogLevelProvider(commOpts),
			"",
		)
		goroutinesNo = 200
		logsNo       = 10
		wg           sync.WaitGroup
		sum          int64
	)
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		var logData map[string]any
		if err := json.Unmarshal(p, &logData); err != nil {
			t.Error(err.Error())

			return 0, err
		}
		assertEqual(t, 4, len(logData))
		assertEqual(t, "bar", logData["foo"].(string))
		assertEqual(t, 10, int(logData["no"].(float64)))
		atomic.AddInt64(&sum, int64(logData["threadNo"].(float64)*logData["logNo"].(float64)))

		return len(p), nil
	})

	// act
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(_ xlog.Formatter, threadNo int) {
			defer wg.Done()
			for j := 0; j < logsNo; j++ {
				keyValues := getInputKeyValues()
				keyValues = append(keyValues, "threadNo", threadNo+1, "logNo", j+1)
				resultErr := subject(writer, keyValues)
				assertNil(t, resultErr)
			}
		}(subject, i)
	}
	wg.Wait()

	// assert
	assertEqual(t, goroutinesNo*logsNo, writer.WriteCallsCount())
	expectedSum := goroutinesNo * (goroutinesNo + 1) * logsNo * (logsNo + 1) / 4
	assertEqual(t, expectedSum, int(sum))
}

func BenchmarkSyslogFormatter_json_syncLogger(b *testing.B) {
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)
	writer := NopSyslogWriter{}
	logger := xlog.NewSyncLogger(
		writer,
		xlog.SyncLoggerWithOptions(commonOpts),
		xlog.SyncLoggerWithFormatter(xlog.SyslogFormatter(
			xlog.JSONFormatter,
			xlog.NewDefaultSyslogLevelProvider(commonOpts),
			"",
		)),
	)
	defer func() {
		_ = logger.Close()
		_ = writer.Close()
	}()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		logger.Error(kv...)
	}
}

func BenchmarkSyslogFormatter_json_asyncLogger(b *testing.B) {
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)
	writer := NopSyslogWriter{}
	logger := xlog.NewAsyncLogger(
		writer,
		xlog.AsyncLoggerWithOptions(commonOpts),
		xlog.AsyncLoggerWithFormatter(xlog.SyslogFormatter(
			xlog.JSONFormatter,
			xlog.NewDefaultSyslogLevelProvider(commonOpts),
			"",
		)),
	)
	defer func() {
		_ = logger.Close()
		_ = writer.Close()
	}()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		logger.Error(kv...)
	}
}

// MockSyslogWriter is a mock for syslogWriter.
type MockSyslogWriter struct {
	*MockWriter
	logCallsCnt   map[syslog.Priority]uint32
	logCallbacks  map[syslog.Priority]func(string) error
	closeCallsCnt uint32
	mu            sync.RWMutex
}

// NewMockSyslogWriter instantiates new mocked syslogWriter.
func NewMockSyslogWriter() *MockSyslogWriter {
	return &MockSyslogWriter{
		MockWriter:   new(MockWriter),
		logCallsCnt:  make(map[syslog.Priority]uint32, 8),
		logCallbacks: make(map[syslog.Priority]func(string) error, 8),
	}
}

// Emerg mock logic.
func (mock *MockSyslogWriter) Emerg(msg string) error {
	return mock.logByLevel(syslog.LOG_EMERG, msg)
}

// Alert mock logic.
func (mock *MockSyslogWriter) Alert(msg string) error {
	return mock.logByLevel(syslog.LOG_ALERT, msg)
}

// Crit mock logic.
func (mock *MockSyslogWriter) Crit(msg string) error {
	return mock.logByLevel(syslog.LOG_CRIT, msg)
}

// Err mock logic.
func (mock *MockSyslogWriter) Err(msg string) error {
	return mock.logByLevel(syslog.LOG_ERR, msg)
}

// Warning mock logic.
func (mock *MockSyslogWriter) Warning(msg string) error {
	return mock.logByLevel(syslog.LOG_WARNING, msg)
}

// Notice mock logic.
func (mock *MockSyslogWriter) Notice(msg string) error {
	return mock.logByLevel(syslog.LOG_NOTICE, msg)
}

// Info mock logic.
func (mock *MockSyslogWriter) Info(msg string) error {
	return mock.logByLevel(syslog.LOG_INFO, msg)
}

// Debug mock logic.
func (mock *MockSyslogWriter) Debug(msg string) error {
	return mock.logByLevel(syslog.LOG_DEBUG, msg)
}

func (mock *MockSyslogWriter) logByLevel(lvl syslog.Priority, msg string) error {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.logCallsCnt[lvl]++

	if mock.logCallbacks[lvl] != nil {
		return mock.logCallbacks[lvl](msg)
	}

	return nil
}

// Close mock logic.
func (mock *MockSyslogWriter) Close() error {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.closeCallsCnt++

	return nil
}

// SetLogCallback sets the callback to be executed inside Emerg/Alert/Crit/Warning/Info/Debug.
// You can make assertions upon passed parameter(s) this way.
func (mock *MockSyslogWriter) SetLogCallback(
	lvl syslog.Priority,
	callback func(msg string) error,
) {
	mock.logCallbacks[lvl] = callback
}

// LogCallsCount returns the no. of times Emerg/Alert/Crit/Warning/Info/Debug was called.
// Differentiate methods calls count by passing appropriate level.
func (mock *MockSyslogWriter) LogCallsCount(lvl syslog.Priority) int {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	return int(mock.logCallsCnt[lvl])
}

// CloseCallsCount returns the no. of times Close was called.
func (mock *MockSyslogWriter) CloseCallsCount() int {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	return int(mock.closeCallsCnt)
}

// NopSyslogWriter is a no-operation syslogWriter.
type NopSyslogWriter struct{}

func (NopSyslogWriter) Emerg(string) error {
	return nil
}
func (NopSyslogWriter) Alert(string) error {
	return nil
}
func (NopSyslogWriter) Crit(string) error {
	return nil
}
func (NopSyslogWriter) Err(string) error {
	return nil
}
func (NopSyslogWriter) Warning(string) error {
	return nil
}
func (NopSyslogWriter) Notice(string) error {
	return nil
}
func (NopSyslogWriter) Info(string) error {
	return nil
}
func (NopSyslogWriter) Debug(string) error {
	return nil
}
func (NopSyslogWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (NopSyslogWriter) Close() error {
	return nil
}
