// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/go-logfmt/logfmt"

	"github.com/actforgood/xlog"
)

func ExampleSyncLogger_withLogfmt() {
	// In this example we create a SyncLogger that writes logs
	// in logfmt format.

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelDebug)
	opts.AdditionalKeyValues = []interface{}{
		"appName", "demo",
		"env", "dev",
	}
	opts.Time = func() interface{} { // mock time for output check
		return "2022-04-12T16:01:20Z"
	}
	opts.Source = xlog.SourceProvider(4, 1) // keep only filename for output check
	logger := xlog.NewSyncLogger(
		os.Stdout,
		xlog.SyncLoggerWithOptions(opts),
		xlog.SyncLoggerWithFormatter(xlog.LogfmtFormatter),
	)
	defer logger.Close()

	logger.Info(xlog.MessageKey, "Hello World", "year", 2022)

	// Output:
	// date=2022-04-12T16:01:20Z lvl=INFO src=/formatter_logfmt_test.go:43 appName=demo env=dev msg="Hello World" year=2022
}

func TestLogfileFormatter_successfullyWritesKeyValues(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.LogfmtFormatter
		dummy     = dummyStringer{Name: "John Doe"}
		someErr   = errors.New("test err.Error() is serialized")
		keyValues = []interface{}{
			"foo", "bar",
			"age", 34,
			"computation", 123.456,
			10, "ten",
			"ints-slice",
			[]int{1, 2, 3},
			dummy, dummy,
			"err", someErr,
		}
		writer bytes.Buffer
	)

	// act
	resultErr := subject(&writer, keyValues)

	// assert
	assertNil(t, resultErr)
	kvMap := make(map[string]string, len(keyValues))
	dec := logfmt.NewDecoder(&writer)
	linesCount := 0
	for dec.ScanRecord() {
		for dec.ScanKeyval() {
			kvMap[string(dec.Key())] = string(dec.Value())
		}
		linesCount++
	}
	if dec.Err() != nil {
		t.Fatal(dec.Err())
	}
	assertEqual(t, 7, len(kvMap))
	assertEqual(t, "bar", kvMap["foo"])
	assertEqual(t, "34", kvMap["age"])
	assertEqual(t, "123.456", kvMap["computation"])
	assertEqual(t, "ten", kvMap["10"])
	assertEqual(t, logfmt.ErrUnsupportedValueType.Error(), kvMap["ints-slice"])
	assertEqual(t, "dummyStringer: John Doe", kvMap["dummyStringer:JohnDoe"])
	assertEqual(t, someErr.Error(), kvMap["err"])
	assertEqual(t, 1, linesCount)
}

func TestLogfmtFormatter_returnsWriteErr(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.LogfmtFormatter
		dummy     = dummyStringer{Name: "John Doe"}
		keyValues = []interface{}{
			"foo", "bar",
			"age", 34,
			"computation", 123.456,
			10, "ten",
			"ints-slice",
			[]int{1, 2, 3},
			dummy, dummy,
		}
		writer = new(MockWriter)
	)
	writer.SetWriteCallback(WriteCallbackErr)

	// act
	resultErr := subject(writer, keyValues)

	// assert
	assertNotNil(t, resultErr)
	assertTrue(t, errors.Is(resultErr, ErrWrite))
}

func BenchmarkLogfmtFormatter(b *testing.B) {
	var (
		subject = xlog.LogfmtFormatter
		dummy   = dummyStringer{Name: "John Doe"}
		input   = []interface{}{
			"foo", "bar",
			"age", 34,
			"computation", 123.456,
			10, "ten",
			"ints-slice",
			[]int{1, 2, 3},
			dummy, dummy,
		}
	)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = subject(io.Discard, input)
	}
}

func TestLogfmtFormatter_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		goroutinesNo = 200
		logsNo       = 10
		wg           sync.WaitGroup
		writer       bytes.Buffer
		sw           = xlog.NewSyncWriter(&writer)
		subject      = xlog.LogfmtFormatter
	)

	// act
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(threadNo int) {
			defer wg.Done()
			for j := 0; j < logsNo; j++ {
				keyValues := getInputKeyValues()
				keyValues = append(keyValues, "threadNo", threadNo+1, "logNo", j+1)
				resultErr := subject(sw, keyValues)
				assertNil(t, resultErr)
			}
		}(i)
	}
	wg.Wait()

	// assert
	var linesCount, sum int
	dec := logfmt.NewDecoder(&writer)
	for dec.ScanRecord() {
		logData := make(map[string]string, len(getInputKeyValues())/2+2)
		for dec.ScanKeyval() {
			logData[string(dec.Key())] = string(dec.Value())
		}
		linesCount++

		assertEqual(t, 4, len(logData))
		assertEqual(t, "bar", logData["foo"])
		assertEqual(t, "10", logData["no"])
		threadNo, _ := strconv.Atoi(logData["threadNo"])
		logNo, _ := strconv.Atoi(logData["logNo"])
		sum += threadNo * logNo
	}
	if dec.Err() != nil {
		t.Fatal(dec.Err())
	}

	assertEqual(t, goroutinesNo*logsNo, linesCount)
	expectedSum := goroutinesNo * (goroutinesNo + 1) * logsNo * (logsNo + 1) / 4
	assertEqual(t, expectedSum, sum)
}
