// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/actforgood/xlog"

	"github.com/getsentry/sentry-go"
)

func ExampleSyncLogger_withSentry() {
	// In this example we create a SyncLogger that sends logs to Sentry.

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         "https://examplePublicKey@o0.ingest.sentry.io/0",
		Environment: "dev",
		Release:     "demo@1.0.0",
	})
	if err != nil {
		panic(err)
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("appName", "demo")
	})

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelDebug)
	logger := xlog.NewSyncLogger(
		io.Discard,
		xlog.SyncLoggerWithOptions(opts),
		xlog.SyncLoggerWithFormatter(xlog.SentryFormatter(
			xlog.JSONFormatter,
			sentry.CurrentHub().Clone(),
			opts,
		)),
	)
	defer func() {
		logger.Close()
		_ = sentry.Flush(2 * time.Second)
	}()

	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)
}

func TestSentryFormatter_successfullySendsDataToSentry(t *testing.T) {
	t.Parallel()

	// arranges
	var (
		sentryHub       = setUpSentryHub()
		commOpts        = xlog.NewCommonOpts()
		formatter       = new(MockFormatter)
		expectedMessage = "foo"
		levels          = []xlog.Level{
			xlog.LevelNone,
			xlog.LevelDebug,
			xlog.LevelInfo,
			xlog.LevelWarning,
			xlog.LevelError,
			xlog.LevelCritical,
		}
		extractedSentryLevel          sentry.Level
		scopeMessageProcessedCallsCnt = 0
	)
	sentryHub.Scope().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		scopeMessageProcessedCallsCnt++
		if assertNotNil(t, event) {
			assertEqual(t, expectedMessage, event.Message)
			extractedSentryLevel = event.Level
		}

		return event
	})

	for idx, level := range levels {
		testLvl := level // capture range variable
		testIdx := idx
		t.Run(fmt.Sprintf("level_%s", commOpts.LevelLabels[testLvl]), func(t *testing.T) {
			keyValues := getExpectedKeyValues(testLvl, commOpts.LevelLabels)
			subject := xlog.SentryFormatter(
				formatter.Format,
				sentryHub,
				commOpts,
			)
			formatter.SetFormatCallback(func(w io.Writer, kv []interface{}) error {
				assertEqual(t, keyValues, kv)
				buf, ok := w.(*bytes.Buffer)
				if assertTrue(t, ok) {
					_, _ = buf.Write([]byte(expectedMessage))
				}

				return nil
			})

			// act
			resultErr := subject(io.Discard, keyValues)

			// assert
			assertNil(t, resultErr)
			assertEqual(t, testIdx+1, formatter.FormatCallsCount())
			assertEqual(t, testIdx+1, scopeMessageProcessedCallsCnt)
			switch testLvl {
			case xlog.LevelDebug:
				assertEqual(t, sentry.LevelDebug, extractedSentryLevel)
			case xlog.LevelWarning:
				assertEqual(t, sentry.LevelWarning, extractedSentryLevel)
			case xlog.LevelError:
				assertEqual(t, sentry.LevelError, extractedSentryLevel)
			case xlog.LevelCritical:
				assertEqual(t, sentry.LevelFatal, extractedSentryLevel)
			default:
				assertEqual(t, sentry.LevelInfo, extractedSentryLevel)
			}
		})
	}
}

func TestSentryFormatter_returnsErrFromFormatter(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		sentryHub = setUpSentryHub()
		commOpts  = xlog.NewCommonOpts()
		formatter = new(MockFormatter)
		subject   = xlog.SentryFormatter(
			formatter.Format,
			sentryHub,
			commOpts,
		)
		keyValues                     = getExpectedKeyValues(xlog.LevelInfo, commOpts.LevelLabels)
		scopeMessageProcessedCallsCnt = 0
	)
	sentryHub.Scope().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		scopeMessageProcessedCallsCnt++

		return event
	})
	formatter.SetFormatCallback(FormatCallbackErr)

	// act
	resultErr := subject(io.Discard, keyValues)

	// assert
	assertTrue(t, errors.Is(resultErr, ErrFormat))
	assertEqual(t, 1, formatter.FormatCallsCount())
	assertEqual(t, 0, scopeMessageProcessedCallsCnt)
}

func TestSentryFormatter_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		sentryHub = setUpSentryHub()
		commOpts  = xlog.NewCommonOpts()
		subject   = xlog.SentryFormatter(
			xlog.JSONFormatter,
			sentryHub,
			commOpts,
		)
		goroutinesNo                  = 200
		logsNo                        = 10
		wg                            sync.WaitGroup
		scopeMessageProcessedCallsCnt = 0
		sum                           = 0
	)
	sentryHub.Scope().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		scopeMessageProcessedCallsCnt++
		var logData map[string]interface{}
		if err := json.Unmarshal([]byte(event.Message), &logData); err != nil {
			t.Error(err.Error())

			return event
		}
		assertEqual(t, 4, len(logData))
		assertEqual(t, "bar", logData["foo"].(string))
		assertEqual(t, 10, int(logData["no"].(float64)))
		sum += int(logData["threadNo"].(float64) * logData["logNo"].(float64))

		return event
	})

	// act
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(logger xlog.Formatter, threadNo int) {
			defer wg.Done()
			for j := 0; j < logsNo; j++ {
				keyValues := getInputKeyValues()
				keyValues = append(keyValues, "threadNo", threadNo+1, "logNo", j+1)
				resultErr := subject(io.Discard, keyValues)
				assertNil(t, resultErr)
			}
		}(subject, i)
	}
	wg.Wait()

	// assert
	assertEqual(t, goroutinesNo*logsNo, scopeMessageProcessedCallsCnt)
	expectedSum := goroutinesNo * (goroutinesNo + 1) * logsNo * (logsNo + 1) / 4
	assertEqual(t, expectedSum, sum)
}

func BenchmarkSentryFormatter_json_syncLogger(b *testing.B) {
	sentryHub := setUpSentryHub()
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)
	logger := xlog.NewSyncLogger(
		io.Discard,
		xlog.SyncLoggerWithOptions(commonOpts),
		xlog.SyncLoggerWithFormatter(xlog.SentryFormatter(
			xlog.JSONFormatter,
			sentryHub,
			commonOpts,
		)),
	)
	defer func() {
		logger.Close()
		_ = sentryHub.Flush(2 * time.Second)
	}()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		logger.Error(kv...)
	}
}

func BenchmarkSentryFormatter_json_asyncLogger(b *testing.B) {
	sentryHub := setUpSentryHub()
	commonOpts := xlog.NewCommonOpts()
	commonOpts.Source = xlog.SourceProvider(4, 1)
	logger := xlog.NewAsyncLogger(
		io.Discard,
		xlog.AsyncLoggerWithOptions(commonOpts),
		xlog.AsyncLoggerWithFormatter(xlog.SentryFormatter(
			xlog.JSONFormatter,
			sentryHub,
			commonOpts,
		)),
	)
	defer func() {
		logger.Close()
		_ = sentryHub.Flush(2 * time.Second)
	}()
	kv := getBenchmarkKeyVals()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		logger.Error(kv...)
	}
}

func setUpSentryHub() *sentry.Hub {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         "", // disables sentry transport
		Release:     "xlog@0.0.1-test",
		Environment: "test",
	})
	if err != nil {
		panic(err)
	}
	scope := sentry.NewScope()
	hub := sentry.NewHub(client, scope)

	return hub
}
