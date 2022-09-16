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
	"testing"

	"github.com/actforgood/xlog"
)

func ExampleSyncLogger_devLogger() {
	// In this example we create a "dev", colorized levels, logger based on TextFormatter.

	opts := xlog.NewCommonOpts()
	opts.MinLevel = xlog.FixedLevelProvider(xlog.LevelNone)
	opts.Time = func() interface{} { // mock time for output check
		return "2022-03-14T16:01:20Z"
	}
	opts.Source = xlog.SourceProvider(4, 1)   // keep only filename for output check
	opts.LevelLabels = map[xlog.Level]string{ // we nicely colorize the levels
		xlog.LevelDebug:    "\033[0;34mDEBUG\033[0m",    // blue
		xlog.LevelInfo:     "\033[0;36mINFO\033[0m",     // cyan
		xlog.LevelWarning:  "\033[0;33mWARN\033[0m",     // yellow
		xlog.LevelError:    "\033[0;31mERROR\033[0m",    // red
		xlog.LevelCritical: "\033[0;31mCRITICAL\033[0m", // red
	}
	logger := xlog.NewSyncLogger(
		os.Stdout,
		xlog.SyncLoggerWithOptions(opts),
		xlog.SyncLoggerWithFormatter(xlog.TextFormatter(opts)),
	)
	defer logger.Close()

	logger.Debug(xlog.MessageKey, "Hello World", "year", 2022)

	// Output:
	// 2022-03-14T16:01:20Z /formatter_text_test.go:41 [0;34mDEBUG[0m Hello World year=2022
}

func TestTextFormatter_successfullyWritesText(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.TextFormatter(xlog.NewCommonOpts())
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
			"lvl", "DEBUG",
			"date", "2021-11-30T16:01:20Z",
			"msg", "Hello World",
			"src", "/formatter_text_test.go:30",
			"err", someErr,
		}
		writer         bytes.Buffer
		expectedResult = "2021-11-30T16:01:20Z /formatter_text_test.go:30 DEBUG Hello World foo=bar age=34 computation=123.456 10=ten ints-slice=[1 2 3] dummyStringer: John Doe=dummyStringer: John Doe err=test err.Error() is serialized\n"
	)

	// act
	resultErr := subject(&writer, keyValues)

	// assert
	assertNil(t, resultErr)
	writtenBytes := writer.Bytes()
	assertEqual(t, expectedResult, string(writtenBytes))
}

func TestTextFormatter_returnsWriteErr(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.TextFormatter(xlog.NewCommonOpts())
		keyValues = []interface{}{"foo", "bar"}
		writer    = new(MockWriter)
	)
	writer.SetWriteCallback(WriteCallbackErr)

	// act
	resultErr := subject(writer, keyValues)

	// assert
	assertNotNil(t, resultErr)
	assertTrue(t, errors.Is(resultErr, ErrWrite))
}

func BenchmarkTextFormatter(b *testing.B) {
	var (
		subject = xlog.TextFormatter(xlog.NewCommonOpts())
		dummy   = dummyStringer{Name: "John Doe"}
		input   = []interface{}{
			"foo", "bar",
			"age", 34,
			"computation", 123.456,
			10, "ten",
			"ints-slice",
			[]int{1, 2, 3},
			dummy, dummy,
			"lvl", "ERR",
			"date", "2021-11-30T16:01:20Z",
			"msg", "Hello World",
			"src", "/formatter_text_test.go:76",
		}
	)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = subject(io.Discard, input)
	}
}
