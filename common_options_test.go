// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/actforgood/xlog"
)

// timeBuffer is a buffer to take around a Time call checks.
const timeBuffer = 200 * time.Millisecond

func TestNewCommonOpts(t *testing.T) {
	t.Parallel()

	subject := xlog.NewCommonOpts()

	t.Run("default level options", func(t *testing.T) {
		t.Parallel()
		if assertNotNil(t, subject.MinLevel) {
			assertEqual(t, xlog.LevelWarning, subject.MinLevel())
		}
		if assertNotNil(t, subject.MaxLevel) {
			assertEqual(t, xlog.LevelCritical, subject.MaxLevel())
		}

		if assertNotNil(t, subject.LevelLabels) {
			assertEqual(t, 5, len(subject.LevelLabels))
			assertEqual(t, "CRITICAL", subject.LevelLabels[xlog.LevelCritical])
			assertEqual(t, "ERROR", subject.LevelLabels[xlog.LevelError])
			assertEqual(t, "WARN", subject.LevelLabels[xlog.LevelWarning])
			assertEqual(t, "INFO", subject.LevelLabels[xlog.LevelInfo])
			assertEqual(t, "DEBUG", subject.LevelLabels[xlog.LevelDebug])
			assertEqual(t, "", subject.LevelLabels[xlog.LevelNone])
		}

		assertEqual(t, "lvl", subject.LevelKey)
	})

	t.Run("default time options", func(t *testing.T) {
		t.Parallel()
		assertEqual(t, "date", subject.TimeKey)

		if assertNotNil(t, subject.Time) {
			before := time.Now().UTC().Add(-1 * timeBuffer)
			result := subject.Time()
			after := time.Now().UTC().Add(timeBuffer)
			checkTime(t, result, before, after, time.RFC3339Nano)
		}
	})

	t.Run("default source options", func(t *testing.T) {
		t.Parallel()
		assertEqual(t, "src", subject.SourceKey)

		if assertNotNil(t, subject.Source) {
			assertTrue(t, subject.Source() == "")
			subject.Source = xlog.SourceProvider(1, 0)
			assertFalse(t, subject.Source() == "")
		}
	})

	t.Run("default additional key values options", func(t *testing.T) {
		t.Parallel()
		assertNil(t, subject.AdditionalKeyValues)
	})

	t.Run("default err handler option", func(t *testing.T) {
		t.Parallel()
		assertNotNil(t, subject.ErrHandler)
	})
}

func TestCommonOpts_BetweenMinMax(t *testing.T) {
	t.Parallel()

	t.Run("none", testCommonOptsBetweenMinMaxLevelNone)
	t.Run("debug", testCommonOptsBetweenMinMaxLevelDebug)
	t.Run("info", testCommonOptsBetweenMinMaxLevelInfo)
	t.Run("warning", testCommonOptsBetweenMinMaxLevelWarning)
	t.Run("error", testCommonOptsBetweenMinMaxLevelError)
	t.Run("critical", testCommonOptsBetweenMinMaxLevelCritical)
}

func testCommonOptsBetweenMinMaxLevelNone(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between None and None",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelNone,
		},
		{
			name:     "Between None and Critical",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between Debug and Critical",
			expected: false,
			min:      xlog.LevelDebug,
			max:      xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelNone)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func testCommonOptsBetweenMinMaxLevelDebug(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between Debug and Debug",
			expected: true,
			min:      xlog.LevelDebug,
			max:      xlog.LevelDebug,
		},
		{
			name:     "Between None and Debug",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelDebug,
		},
		{
			name:     "Between Debug and Critical",
			expected: true,
			min:      xlog.LevelDebug,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between None and None",
			expected: false,
			min:      xlog.LevelNone,
			max:      xlog.LevelNone,
		},
		{
			name:     "Not Between Warning and Critical",
			expected: false,
			min:      xlog.LevelWarning,
			max:      xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelDebug)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func testCommonOptsBetweenMinMaxLevelInfo(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between Info and Info",
			expected: true,
			min:      xlog.LevelInfo,
			max:      xlog.LevelInfo,
		},
		{
			name:     "Between None and Info",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelInfo,
		},
		{
			name:     "Between Info and Critical",
			expected: true,
			min:      xlog.LevelInfo,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between None and Debug",
			expected: false,
			min:      xlog.LevelNone,
			max:      xlog.LevelDebug,
		},
		{
			name:     "Not Between Warning and Critical",
			expected: false,
			min:      xlog.LevelWarning,
			max:      xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelInfo)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func testCommonOptsBetweenMinMaxLevelWarning(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between Warning and Warning",
			expected: true,
			min:      xlog.LevelWarning,
			max:      xlog.LevelWarning,
		},
		{
			name:     "Between None and Warning",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelWarning,
		},
		{
			name:     "Between Warning and Critical",
			expected: true,
			min:      xlog.LevelWarning,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between None and Info",
			expected: false,
			min:      xlog.LevelNone,
			max:      xlog.LevelInfo,
		},
		{
			name:     "Not Between Error and Critical",
			expected: false,
			min:      xlog.LevelError,
			max:      xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelWarning)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func testCommonOptsBetweenMinMaxLevelError(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between Error and Error",
			expected: true,
			min:      xlog.LevelError,
			max:      xlog.LevelError,
		},
		{
			name:     "Between None and Error",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelError,
		},
		{
			name:     "Between Error and Critical",
			expected: true,
			min:      xlog.LevelError,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between None and Warning",
			expected: false,
			min:      xlog.LevelNone,
			max:      xlog.LevelWarning,
		},
		{
			name:     "Not Between Critical and Critical",
			expected: false,
			min:      xlog.LevelCritical,
			max:      xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelError)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func testCommonOptsBetweenMinMaxLevelCritical(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected bool
		min      xlog.Level
		max      xlog.Level
	}{
		{
			name:     "Between Critical and Critical",
			expected: true,
			min:      xlog.LevelCritical,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Between None and Critical",
			expected: true,
			min:      xlog.LevelNone,
			max:      xlog.LevelCritical,
		},
		{
			name:     "Not Between None and Error",
			expected: false,
			min:      xlog.LevelNone,
			max:      xlog.LevelError,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			subject.MinLevel = xlog.FixedLevelProvider(test.min)
			subject.MaxLevel = xlog.FixedLevelProvider(test.max)

			// act
			result := subject.BetweenMinMax(xlog.LevelCritical)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func TestCommonOpts_WithDefaultKeyValues(t *testing.T) {
	t.Parallel()

	t.Run("time, level, source", testCommonOptsDefaultKeyValuesTimeLevelSource)
	t.Run("time, level, source, additional key values", testCommonOptsDefaultKeyValuesTimeLevelSourceAdditionalKeyValues)
	t.Run("no source (key)", testCommonOptsDefaultKeyValuesNoSourceKey)
	t.Run("no source (value)", testCommonOptsDefaultKeyValuesNoSourceValue)
	t.Run("level", testCommonOptsDefaultKeyValuesLevel)
	t.Run("default with custom", testCommonOptsDefaultKeyValuesWithCustom)
}

func testCommonOptsDefaultKeyValuesTimeLevelSource(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.NewCommonOpts()
		before, after time.Time
	)

	// act
	before = time.Now().UTC().Add(-1 * timeBuffer)
	result := subject.WithDefaultKeyValues(xlog.LevelError)
	after = time.Now().UTC().Add(timeBuffer)

	// assert
	if !assertEqual(t, 6, len(result)) {
		t.FailNow()
	}

	assertEqual(t, "date", result[0])
	checkTime(t, result[1], before, after, time.RFC3339Nano)

	assertEqual(t, "lvl", result[2])
	assertEqual(t, "ERROR", result[3])

	assertEqual(t, "src", result[4])
	assertFalse(t, result[5] == "")
}

func testCommonOptsDefaultKeyValuesTimeLevelSourceAdditionalKeyValues(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.NewCommonOpts()
		before, after time.Time
	)
	subject.AdditionalKeyValues = []interface{}{
		"foo", "bar",
		"provider-key", xlog.Provider(func() interface{} {
			return "provider-value"
		}),
	}

	// act
	before = time.Now().UTC().Add(-1 * timeBuffer)
	result := subject.WithDefaultKeyValues(xlog.LevelError)
	after = time.Now().UTC().Add(timeBuffer)

	// assert
	if !assertEqual(t, 10, len(result)) {
		t.FailNow()
	}

	assertEqual(t, "date", result[0])
	checkTime(t, result[1], before, after, time.RFC3339Nano)

	assertEqual(t, "lvl", result[2])
	assertEqual(t, "ERROR", result[3])

	assertEqual(t, "src", result[4])
	assertFalse(t, result[5] == "")

	assertEqual(t, "foo", result[6])
	assertEqual(t, "bar", result[7])
	assertEqual(t, "provider-key", result[8])
	assertEqual(t, "provider-value", result[9])
}

func testCommonOptsDefaultKeyValuesNoSourceKey(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.NewCommonOpts()
		before, after time.Time
	)
	subject.SourceKey = ""

	// act
	before = time.Now().UTC().Add(-1 * timeBuffer)
	result := subject.WithDefaultKeyValues(xlog.LevelError)
	after = time.Now().UTC().Add(timeBuffer)

	// assert
	if !assertEqual(t, 4, len(result)) {
		t.FailNow()
	}

	assertEqual(t, "date", result[0])
	checkTime(t, result[1], before, after, time.RFC3339Nano)

	assertEqual(t, "lvl", result[2])
	assertEqual(t, "ERROR", result[3])
}

func testCommonOptsDefaultKeyValuesNoSourceValue(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.NewCommonOpts()
		before, after time.Time
	)
	subject.Source = xlog.SourceProvider(200, 0) // too many frames to skip

	// act
	before = time.Now().UTC().Add(-1 * timeBuffer)
	result := subject.WithDefaultKeyValues(xlog.LevelError)
	after = time.Now().UTC().Add(timeBuffer)

	// assert
	if !assertEqual(t, 4, len(result)) {
		t.FailNow()
	}

	assertEqual(t, "date", result[0])
	checkTime(t, result[1], before, after, time.RFC3339Nano)

	assertEqual(t, "lvl", result[2])
	assertEqual(t, "ERROR", result[3])
}

func testCommonOptsDefaultKeyValuesLevel(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xlog.NewCommonOpts()
	tests := [...]struct {
		name     string
		expected string
		input    xlog.Level
	}{
		{
			name:     "LevelKey is not present for None",
			expected: "",
			input:    xlog.LevelNone,
		},
		{
			name:     "LevelKey is present for Debug",
			expected: "DEBUG",
			input:    xlog.LevelDebug,
		},
		{
			name:     "LevelKey is present for Info",
			expected: "INFO",
			input:    xlog.LevelInfo,
		},
		{
			name:     "LevelKey is present for Warning",
			expected: "WARN",
			input:    xlog.LevelWarning,
		},
		{
			name:     "LevelKey is present for Error",
			expected: "ERROR",
			input:    xlog.LevelError,
		},
		{
			name:     "LevelKey is present for Critical",
			expected: "CRITICAL",
			input:    xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := subject.WithDefaultKeyValues(test.input)

			// assert
			foundLevelKey := false
			foundLevelValue := ""
			for idx, kv := range result {
				if idx%2 == 0 && kv == "lvl" {
					foundLevelKey = true
					foundLevelValue, _ = result[idx+1].(string)

					break
				}
			}
			if test.expected == "" {
				assertFalse(t, foundLevelKey)
			} else {
				assertTrue(t, foundLevelKey)
				assertEqual(t, test.expected, foundLevelValue)
			}
		})
	}
}

func testCommonOptsDefaultKeyValuesWithCustom(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.NewCommonOpts()
		before, after time.Time
		keyValues     = []interface{}{"a", "b", "one", 1}
	)

	// act
	before = time.Now().UTC().Add(-1 * timeBuffer)
	result := subject.WithDefaultKeyValues(xlog.LevelError, keyValues...)
	after = time.Now().UTC().Add(timeBuffer)

	// assert
	if !assertEqual(t, 10, len(result)) {
		t.FailNow()
	}

	assertEqual(t, "date", result[0])
	checkTime(t, result[1], before, after, time.RFC3339Nano)

	assertEqual(t, "lvl", result[2])
	assertEqual(t, "ERROR", result[3])

	assertEqual(t, "src", result[4])
	assertFalse(t, result[5] == "")

	assertEqual(t, "a", result[6])
	assertEqual(t, "b", result[7])
	assertEqual(t, "one", result[8])
	assertEqual(t, 1, result[9])
}

func TestFixedLevelProvider(t *testing.T) {
	t.Parallel()

	// arrange
	tests := [...]struct {
		name  string
		input xlog.Level
	}{
		{
			name:  "None",
			input: xlog.LevelNone,
		},
		{
			name:  "Debug",
			input: xlog.LevelDebug,
		},
		{
			name:  "Info",
			input: xlog.LevelInfo,
		},
		{
			name:  "Warning",
			input: xlog.LevelWarning,
		},
		{
			name:  "Error",
			input: xlog.LevelError,
		},
		{
			name:  "Critical",
			input: xlog.LevelCritical,
		},
	}

	for _, testData := range tests {
		data := testData // capture range variable
		t.Run(data.name, func(t *testing.T) {
			t.Parallel()

			subject := xlog.FixedLevelProvider(data.input)
			for i := 0; i < 3; i++ {
				// act
				result := subject()

				// assert
				assertEqual(t, data.input, result)
			}
		})
	}
}

func TestEnvLevelProvider(t *testing.T) {
	t.Parallel()

	t.Run("valid env", testEnvLevelProviderWithValidEnv)
	t.Run("invalid env", testEnvLevelProviderWithInvalidEnv)
	t.Run("not found env", testEnvLevelProviderWithNotFoundEnv)
}

func testEnvLevelProviderWithValidEnv(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject     = xlog.EnvLevelProvider
		lvl         = xlog.LevelDebug
		defaultLvl  = xlog.LevelInfo
		envName     = setUpLevelEnv("DEBUG")
		levelLabels = map[xlog.Level]string{lvl: "DEBUG", xlog.LevelWarning: "WARN"}
	)
	if envName == "" {
		t.Fatal("could not setup level env")
	}
	defer tearDownLevelEnv(envName)

	// act
	result1 := subject(envName, defaultLvl, levelLabels)()
	result2 := subject(envName, defaultLvl, levelLabels)()

	// assert
	assertEqual(t, lvl, result1)
	assertEqual(t, result1, result2)

	// change the value and see new value is returned.
	newLvl := xlog.LevelWarning
	err := os.Setenv(envName, "WARN")
	if !assertNil(t, err) {
		t.FailNow()
	}

	// act
	result3 := subject(envName, defaultLvl, levelLabels)()
	result4 := subject(envName, defaultLvl, levelLabels)()

	// assert
	assertEqual(t, newLvl, result3)
	assertEqual(t, result3, result4)
}

func testEnvLevelProviderWithInvalidEnv(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject     = xlog.EnvLevelProvider
		defaultLvl  = xlog.LevelInfo
		envName     = setUpLevelEnv("unknown")
		levelLabels = map[xlog.Level]string{xlog.LevelWarning: "WARN"}
	)
	if envName == "" {
		t.Fatal("could not setup level env")
	}
	defer tearDownLevelEnv(envName)

	// act
	result := subject(envName, defaultLvl, levelLabels)()

	// assert
	assertEqual(t, defaultLvl, result)
}

func testEnvLevelProviderWithNotFoundEnv(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject     = xlog.EnvLevelProvider
		defaultLvl  = xlog.LevelInfo
		envName     = "MY_NOT_FOUND_LOG_LEVEL_24fa58a5-544e-4f89-9828-73d49962de15"
		levelLabels = map[xlog.Level]string{xlog.LevelWarning: "WARN"}
	)
	if _, found := os.LookupEnv(envName); found {
		t.Skip("OS env exists")
	}

	// act
	result := subject(envName, defaultLvl, levelLabels)()

	// assert
	assertEqual(t, defaultLvl, result)
}

func TestUTCTimeProvider(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject = xlog.LocalTimeProvider
		format  = "2006-01-02"
	)

	// act
	before, _ := time.Parse(format, time.Now().Format(format))
	before = before.Add(-1 * timeBuffer).UTC()
	result := subject(format)()
	after, _ := time.Parse(format, time.Now().Format(format))
	after = after.Add(timeBuffer).UTC()

	// assert
	checkTime(t, result, before, after, format)
}

func TestLocalTimeProvider(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject = xlog.LocalTimeProvider
		format  = "2006-01-02"
	)

	// act
	before, _ := time.Parse(format, time.Now().Format(format))
	before = before.Add(-1 * timeBuffer).UTC()
	result := subject(format)()
	after, _ := time.Parse(format, time.Now().Format(format))
	after = after.Add(timeBuffer).UTC()

	// assert
	checkTime(t, result, before, after, format)
}

func TestSourceProvider(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject = xlog.SourceProvider
		reg     = regexp.MustCompile(`common_options_test\.go:\d+$`)
	)

	// act
	result := subject(1, 0)()
	resultStr, ok := result.(string)

	// assert
	if assertTrue(t, ok) {
		assertTrue(t, reg.MatchString(resultStr))
	}

	// act
	result2 := subject(1, 1)() // with 1 path skipped.
	result2Str, ok := result2.(string)

	// assert
	if assertTrue(t, ok) {
		assertTrue(t, reg.MatchString(result2Str))
		assertTrue(t, len(resultStr) > len(result2Str))
	}

	// act
	result3 := subject(2, 0)() // with 2 frames skipped
	result3Str, ok := result3.(string)

	// assert
	if assertTrue(t, ok) {
		assertFalse(t, reg.MatchString(result3Str))
	}
}

func TestAppendNoValue(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject       = xlog.AppendNoValue
		oddKeyValues  = []interface{}{"key"}
		evenKeyValues = []interface{}{"key", "value"}
	)

	// act
	result := subject(oddKeyValues)

	// assert
	assertEqual(t, []interface{}{"key", "*NoValue*"}, result)

	// act
	result2 := subject(evenKeyValues)

	// assert
	assertEqual(t, evenKeyValues, result2)
}

func TestStackErr(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject = xlog.StackErr
		input1  = dummyEnrichedErr{}
		input2  = errors.New("baz")
	)

	// act
	result1 := subject(input1)

	// assert
	assertEqual(t, "foo bar", result1)

	// act
	result2 := subject(input2)

	// assert
	assertEqual(t, "baz", result2)
}

type dummyEnrichedErr struct{}

func (dummyEnrichedErr) Error() string {
	return "dummy err that offers '%+v' support"
}
func (dummyEnrichedErr) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		fmt.Fprint(s, "foo bar")
	}
}

// checkTime is used internally to test a Time result.
func checkTime(t *testing.T, date interface{}, before, after time.Time, format string) {
	t.Helper()
	dateStr, isString := date.(string)
	if assertTrue(t, isString) {
		dateTime, err := time.Parse(format, dateStr)
		if assertNil(t, err) {
			assertTrue(t, before.Before(dateTime.UTC()))
			assertTrue(t, after.After(dateTime.UTC()))
		}
	}
}

// setUpLevelEnv sets OS env with provided level.
// Returns the env name.
// Can be empty if op did not succeed.
func setUpLevelEnv(lvlLabel string) string {
	nBig, err := rand.Int(rand.Reader, big.NewInt(9999999))
	if err != nil {
		return ""
	}
	randInt := nBig.Int64()
	envName := "TEST_LOG_LEVEL_ENV_" + strconv.FormatInt(randInt, 10)
	if _, found := os.LookupEnv(envName); found {
		return ""
	}
	if err := os.Setenv(envName, lvlLabel); err != nil {
		return ""
	}

	return envName
}

// tearDownLevelEnv unsets the OS env provided.
func tearDownLevelEnv(envName string) {
	_ = os.Unsetenv(envName)
}

func BenchmarkCommonOpts_WithDefaultKeyValues(b *testing.B) {
	subject := xlog.NewCommonOpts()
	subject.AdditionalKeyValues = []interface{}{
		"foo", "bar",
		"provider-key", xlog.Provider(func() interface{} {
			return "provider-value"
		}),
		"some-int", 567,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = subject.WithDefaultKeyValues(xlog.LevelInfo)
	}
}
