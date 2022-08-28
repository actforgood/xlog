// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

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

func ExampleAsyncLogger() {
	// In this example we create a (async)logger that writes
	// logs to standard output.

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	opts.AdditionalKeyValues = []interface{}{
		"appName", "demo",
		"env", "dev",
	}
	opts.Time = func() interface{} { // mock time for output check
		return "2022-03-16T16:01:20Z"
	}
	opts.Source = xlog.SourceProvider(4, 1) // keep only filename for output check

	logger := xlog.NewAsyncLogger(
		os.Stdout,
		xlog.AsyncLoggerWithOptions(opts),
		xlog.AsyncLoggerWithWorkersNo(2), // since workers no > 1, we expect output to be unordered.
	)
	defer logger.Close()

	logger.Log(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Debug(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Warn(xlog.MessageKey, "Hello World", "year", 2022)
	logger.Error(xlog.MessageKey, "Could not read file", xlog.ErrorKey, io.ErrUnexpectedEOF, "file", "/some/file")
	logger.Critical(xlog.MessageKey, "DB connection is down")

	// Unordered output:
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","msg":"Hello World","src":"/logger_async_test.go:43","year":2022}
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","lvl":"DEBUG","msg":"Hello World","src":"/logger_async_test.go:44","year":2022}
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","lvl":"INFO","msg":"Hello World","src":"/logger_async_test.go:45","year":2022}
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","lvl":"WARN","msg":"Hello World","src":"/logger_async_test.go:46","year":2022}
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","err":"unexpected EOF","file":"/some/file","lvl":"ERROR","msg":"Could not read file","src":"/logger_async_test.go:47"}
	// {"appName":"demo","date":"2022-03-16T16:01:20Z","env":"dev","lvl":"CRITICAL","msg":"DB connection is down","src":"/logger_async_test.go:48"}
}

func TestAsyncLogger_Log(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelNone
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func TestAsyncLogger_Debug(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelDebug
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func TestAsyncLogger_Info(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelInfo
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func TestAsyncLogger_Warn(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelWarning
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func TestAsyncLogger_Error(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelError
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func TestAsyncLogger_Critical(t *testing.T) {
	t.Parallel()

	testLvl := xlog.LevelCritical
	t.Run("success", testAsyncLoggerLogSuccessful(testLvl))
	t.Run("ignored", testAsyncLoggerLogIgnored(testLvl))
	t.Run("format write err", testAsyncLoggerLogFormatErr(testLvl))
}

func testAsyncLoggerLogSuccessful(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer                 = io.Discard
			formatter              = new(MockFormatter)
			errHandler             = new(MockErrorHandler)
			commOpts               = xlog.NewCommonOpts()
			subject    xlog.Logger = xlog.NewAsyncLogger(
				writer,
				xlog.AsyncLoggerWithFormatter(formatter.Format),
				xlog.AsyncLoggerWithOptions(commOpts),
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

func testAsyncLoggerLogIgnored(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer     = io.Discard
			formatter  = new(MockFormatter)
			errHandler = new(MockErrorHandler)
			commOpts   = xlog.NewCommonOpts()
			subject    = xlog.NewAsyncLogger(
				writer,
				xlog.AsyncLoggerWithFormatter(formatter.Format),
				xlog.AsyncLoggerWithOptions(commOpts),
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

func testAsyncLoggerLogFormatErr(testLvl xlog.Level) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// arrange
		var (
			writer     = io.Discard
			formatter  = new(MockFormatter)
			errHandler = new(MockErrorHandler)
			commOpts   = xlog.NewCommonOpts()
			subject    = xlog.NewAsyncLogger(
				writer,
				xlog.AsyncLoggerWithFormatter(formatter.Format),
				xlog.AsyncLoggerWithOptions(commOpts),
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

func TestAsyncLogger_Close_withBufferedWriter(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer    bytes.Buffer
		bufWriter = xlog.NewBufferedWriter(
			&writer,
			xlog.BufferedWriterWithSize(1024*1024),
			xlog.BufferedWriterWithFlushInterval(0),
		)
		subject = xlog.NewAsyncLogger(
			bufWriter,
			xlog.AsyncLoggerWithChannelSize(1),
		)
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

func TestAsyncLogger_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		errHandler       = new(MockErrorHandler)
		commOpts         = xlog.NewCommonOpts()
		goroutinesNo     = 200
		logsNo           = 10
		wg               sync.WaitGroup
		buf1, buf2       bytes.Buffer
		writer1, writer2 io.Writer = &buf1, xlog.NewSyncWriter(&buf2)
		tests                      = [...]struct {
			name    string
			buf     *bytes.Buffer
			writer  io.Writer
			subject *xlog.AsyncLogger
		}{
			{
				name: "one worker is safe for un-sync writer",
				buf:  &buf1,
				subject: xlog.NewAsyncLogger(
					writer1,
					xlog.AsyncLoggerWithOptions(commOpts),
				),
			},
			{
				name: "more than one worker needs a sync writer",
				buf:  &buf2,
				subject: xlog.NewAsyncLogger(
					writer2,
					xlog.AsyncLoggerWithOptions(commOpts),
					xlog.AsyncLoggerWithWorkersNo(2),
				),
			},
		}
	)
	commOpts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	commOpts.AdditionalKeyValues = getAdditionalKeyValues()
	commOpts.ErrHandler = errHandler.Handle
	commOpts.SourceKey = ""
	commOpts.Time = staticTimeProvider

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
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
				}(test.subject, i)
			}
			wg.Wait()
			_ = test.subject.Close()

			// assert
			var linesCount, sum int
			for {
				line, err := test.buf.ReadBytes('\n')
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
				assertEqual(t, staticTime, logData["date"])
				assertEqual(t, "extraValue", logData["extraKey"])
				assertEqual(t, "bar", logData["foo"])
				assertEqual(t, float64(10), logData["no"])
				sum += int(logData["threadNo"].(float64) * logData["logNo"].(float64))
			}

			assertEqual(t, 0, errHandler.HandleCallsCount())
			assertEqual(t, goroutinesNo*logsNo, linesCount)
			expectedSum := goroutinesNo * (goroutinesNo + 1) * logsNo * (logsNo + 1) / 4
			assertEqual(t, expectedSum, sum)
		})
	}
}

func BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with1Worker_sequential(b *testing.B) {
	subject := makeAsyncLogger(io.Discard, 256, 1)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with1Worker_parallel(b *testing.B) {
	subject := makeAsyncLogger(io.Discard, 256, 1)
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

func BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with4Workers(b *testing.B) {
	subject := makeAsyncLogger(io.Discard, 256, 4)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withFileWriter_with256ChanSize_with1Worker(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(f, 256, 1)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withBufferedFileWriter_with256ChanSize_with1Worker(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(xlog.NewBufferedWriter(f), 256, 1)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withFileWriter_with1024ChanSize_with4Workers(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(f, 1024, 4)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withBufferedFileWriter_with1024ChanSize_with4Workers(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(xlog.NewBufferedWriter(f), 1024, 4)
	defer subject.Close()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		subject.Error(kv...)
	}
}

func BenchmarkAsyncLogger_json_withDiscardWriter_with256ChanSize_with1Worker_withConcurrency10(b *testing.B) {
	subject := makeAsyncLogger(io.Discard, 10, 1)
	defer subject.Close()
	kv := getBenchmarkKeyVals()
	var wg sync.WaitGroup
	goroutinesNo := 10

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		wg.Add(goroutinesNo)
		for i := 0; i < goroutinesNo; i++ {
			go func() {
				defer wg.Done()
				subject.Error(kv...)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAsyncLogger_json_withDiscardWriter_with1024ChanSize_with4Workers_withConcurrency100(b *testing.B) {
	subject := makeAsyncLogger(io.Discard, 1024, 4)
	defer subject.Close()
	kv := getBenchmarkKeyVals()
	var wg sync.WaitGroup
	goroutinesNo := 100

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		wg.Add(goroutinesNo)
		for i := 0; i < goroutinesNo; i++ {
			go func() {
				defer wg.Done()
				subject.Error(kv...)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAsyncLogger_json_withFileWriter_with256ChanSize_with1Worker_withConcurrency10(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(f, 256, 1)
	defer subject.Close()
	kv := getBenchmarkKeyVals()
	var wg sync.WaitGroup
	goroutinesNo := 10

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		wg.Add(goroutinesNo)
		for i := 0; i < goroutinesNo; i++ {
			go func() {
				defer wg.Done()
				subject.Error(kv...)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAsyncLogger_json_withFileWriter_with1024ChanSize_with4Workers_withConcurrency100(b *testing.B) {
	f := setUpFile(b.Name())
	defer tearDownFile(f)
	subject := makeAsyncLogger(f, 1024, 24)
	defer subject.Close()
	kv := getBenchmarkKeyVals()
	var wg sync.WaitGroup
	goroutinesNo := 100

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		wg.Add(goroutinesNo)
		for i := 0; i < goroutinesNo; i++ {
			go func() {
				defer wg.Done()
				subject.Error(kv...)
			}()
		}
		wg.Wait()
	}
}

// makeAsyncLogger creates a new AsyncLogger object.
func makeAsyncLogger(w io.Writer, chanSize uint16, workersNo uint16) *xlog.AsyncLogger {
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)

	return xlog.NewAsyncLogger(
		w,
		xlog.AsyncLoggerWithOptions(commonOpts),
		xlog.AsyncLoggerWithChannelSize(chanSize),
		xlog.AsyncLoggerWithWorkersNo(workersNo),
	)
}
