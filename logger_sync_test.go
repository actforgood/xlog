// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/actforgood/xlog"
)

func ExampleSyncLogger() {
	// In this example we create a (sync)logger that writes
	// logs to standard output.

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	opts.AdditionalKeyValues = []interface{}{
		"appName", "demo",
		"env", "dev",
	}
	opts.Time = func() interface{} { // mock time for output check
		return "2022-03-14T16:01:20Z"
	}
	opts.Source = xlog.SourceProvider(4, 1) // keep only filename for output check

	logger := xlog.NewSyncLogger(
		os.Stdout,
		xlog.SyncLoggerWithOptions(opts),
	)
	defer logger.Close()

	logger.Log(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Debug(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Warn(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Error(xlog.MessageKey, "Could not read file", xlog.ErrorKey, io.ErrUnexpectedEOF, "file", "/some/file")
	logger.Critical(xlog.MessageKey, "DB connection is down")

	// Output:
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","msg":"Hello World","src":"/logger_sync_test.go:42","year":2022}
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","lvl":"DEBUG","msg":"Hello World","src":"/logger_sync_test.go:43","year":2022}
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","lvl":"INFO","msg":"Hello World","src":"/logger_sync_test.go:44","year":2022}
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","lvl":"WARN","msg":"Hello World","src":"/logger_sync_test.go:45","year":2022}
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","err":"unexpected EOF","file":"/some/file","lvl":"ERROR","msg":"Could not read file","src":"/logger_sync_test.go:46"}
	// {"appName":"demo","date":"2022-03-14T16:01:20Z","env":"dev","lvl":"CRITICAL","msg":"DB connection is down","src":"/logger_sync_test.go:47"}
}

func TestSyncLogger_Log(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelNone
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func TestSyncLogger_Debug(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelDebug
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func TestSyncLogger_Info(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelInfo
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func TestSyncLogger_Warn(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelWarning
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func TestSyncLogger_Error(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelError
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func TestSyncLogger_Critical(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelCritical
	t.Run("success", testSyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testSyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testSyncLoggerLogFormatErr(testLvl))
}

func testSyncLoggerLogSuccessful(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer                 = io.Discard
			formatter              = new(MockFormatter)
			errHandler             = new(MockErrorHandler)
			commOpts               = xlog.NewCommonOpts()
			subject    xlog.Logger = xlog.NewSyncLogger(
				writer,
				xlog.SyncLoggerWithFormatter(formatter.Format),
				xlog.SyncLoggerWithOptions(commOpts),
			)
		)
		commOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
		commOpts.MaxLevel = xlog.FixedLevelProvider(xlog.LevelCritical)
		commOpts.AdditionalKeyValues = getAdditionalKeyValues()
		commOpts.ErrHandler = errHandler.Handle
		commOpts.SourceKey = ""
		commOpts.Time = staticTimeProvider

		formatter.SetFormatCallback(func(w io.Writer, kv []interface{}) error {
			assertEqual(t, getExpectedKeyValues(testLvl, commOpts.LevelLabels), kv)
			assertEqual(t, writer, w)

			return nil
		})

		// act
		callMethodByLevel(subject, testLvl)
		_ = subject.Close()

		// assert
		assertEqual(t, 1, formatter.FormatCallsCount())
		assertEqual(t, 0, errHandler.HandleCallsCount())
	}
}

func testSyncLoggerLogIgnored(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer     = io.Discard
			formatter  = new(MockFormatter)
			errHandler = new(MockErrorHandler)
			commOpts   = xlog.NewCommonOpts()
			subject    = xlog.NewSyncLogger(
				writer,
				xlog.SyncLoggerWithFormatter(formatter.Format),
				xlog.SyncLoggerWithOptions(commOpts),
			)
		)
		if testLvl != xlog.LevelError {
			commOpts.MinLevel = xlog.FixedLevelProvider(testLvl + 1)
			commOpts.MaxLevel = xlog.FixedLevelProvider(xlog.LevelCritical)
		} else {
			commOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
			commOpts.MaxLevel = xlog.FixedLevelProvider(testLvl - 1)
		}
		commOpts.ErrHandler = errHandler.Handle

		// act
		callMethodByLevel(subject, testLvl)
		_ = subject.Close()

		// assert
		assertEqual(t, 0, formatter.FormatCallsCount())
		assertEqual(t, 0, errHandler.HandleCallsCount())
	}
}

func testSyncLoggerLogFormatErr(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer     = io.Discard
			formatter  = new(MockFormatter)
			errHandler = new(MockErrorHandler)
			commOpts   = xlog.NewCommonOpts()
			subject    = xlog.NewSyncLogger(
				writer,
				xlog.SyncLoggerWithFormatter(formatter.Format),
				xlog.SyncLoggerWithOptions(commOpts),
			)
		)
		commOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
		commOpts.MaxLevel = xlog.FixedLevelProvider(xlog.LevelCritical)
		commOpts.AdditionalKeyValues = getAdditionalKeyValues()
		commOpts.ErrHandler = errHandler.Handle
		commOpts.SourceKey = ""
		commOpts.Time = staticTimeProvider
		formatter.SetFormatCallback(FormatCallbackErr)
		errHandler.SetHandleCallback(func(err error, keyVals []interface{}) {
			assertTrue(t, errors.Is(err, ErrFormat))
			assertEqual(t, getExpectedKeyValues(testLvl, commOpts.LevelLabels), keyVals)
		})

		// act
		callMethodByLevel(subject, testLvl)
		_ = subject.Close()

		// assert
		assertEqual(t, 1, formatter.FormatCallsCount())
		assertEqual(t, 1, errHandler.HandleCallsCount())
	}
}

func TestSyncLogger_Close_withBufferedWriter(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer    bytes.Buffer
		bufWriter = xlog.NewBufferedWriter(
			&writer,
			xlog.BufferedWriterWithSize(1024*1024),
			xlog.BufferedWriterWithFlushInterval(0),
		)
		subject = xlog.NewSyncLogger(bufWriter)
	)
	subject.Error("msg", "foo bar")

	// act
	_ = subject.Close() // will call Stop on bufWriter, and log gets flushed.

	// assert
	log, err := writer.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	assertTrue(t, strings.Contains(log, "foo bar"))
}

func TestSyncLogger_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer     bytes.Buffer
		syncWriter = xlog.NewSyncWriter(&writer)
		errHandler = new(MockErrorHandler)
		commOpts   = xlog.NewCommonOpts()
		subject    = xlog.NewSyncLogger(
			syncWriter,
			xlog.SyncLoggerWithOptions(commOpts),
		)
		goroutinesNo = 200
		logsNo       = 10
		wg           sync.WaitGroup
	)
	commOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	commOpts.AdditionalKeyValues = getAdditionalKeyValues()
	commOpts.ErrHandler = errHandler.Handle
	commOpts.SourceKey = ""
	commOpts.Time = staticTimeProvider

	// act
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(logger xlog.Logger, threadNo int) {
			defer wg.Done()
			for j := 0; j < logsNo; j++ {
				keyValues := getInputKeyValues()
				keyValues = append(keyValues, "threadNo", threadNo+1, "logNo", j+1)
				logger.Log(keyValues...)
			}
		}(subject, i)
	}
	wg.Wait()
	_ = subject.Close()

	// assert
	var linesCount, sum int
	for {
		line, err := writer.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Error(err.Error())

			continue
		}
		linesCount++

		var logData map[string]interface{}
		if err := json.Unmarshal(line, &logData); err != nil {
			t.Error(err.Error())

			continue
		}
		assertEqual(t, 6, len(logData))
		assertEqual(t, staticTime, logData["date"].(string))
		assertEqual(t, "extraValue", logData["extraKey"].(string))
		assertEqual(t, "bar", logData["foo"].(string))
		assertEqual(t, 10, int(logData["no"].(float64)))
		sum += int(logData["threadNo"].(float64) * logData["logNo"].(float64))
	}

	assertEqual(t, 0, errHandler.HandleCallsCount())
	assertEqual(t, goroutinesNo*logsNo, linesCount)
	expectedSum := goroutinesNo * (goroutinesNo + 1) * logsNo * (logsNo + 1) / 4
	assertEqual(t, expectedSum, sum)
}

func BenchmarkSyncLogger_json_withDiscardWriter_sequential(b *testing.B) {
	subject := makeSyncLogger(io.Discard)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkSyncLogger_json_withDiscardWriter_parallel(b *testing.B) {
	subject := makeSyncLogger(io.Discard)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			subject.Error(kv...)
		}
	})
}

func BenchmarkSyncLogger_json_withFileWriter(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeSyncLogger(f)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkSyncLogger_json_withBufferedFileWriter(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeSyncLogger(xlog.NewBufferedWriter(f))
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

// makeSyncLogger creates a new SyncLogger object.
func makeSyncLogger(w io.Writer) *xlog.SyncLogger {
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)

	return xlog.NewSyncLogger(
		w,
		xlog.SyncLoggerWithOptions(commonOpts),
	)
}
