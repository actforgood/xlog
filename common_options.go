// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

// MessageKey represents the key under which the main message resides.
// You are not obliged to use this key.
const MessageKey = "msg"

// ErrorKey represents the key under which an error resides.
// For a Logger.Error call usually you'll want to log an error.
// You are not obliged to use this key.
const ErrorKey = "err"

const (
	defaultOptTimeKey   = "date"
	defaultOptLevelKey  = "lvl"
	defaultOptSourceKey = "src"
)

// noValue is a value to be added to key-values logs
// slice in case provided slice is odd.
const noValue = "*NoValue*"

// CommonOpts is a struct holding common configurations for a logger.
type CommonOpts struct {
	// MinLevel is a function that returns the minimum level
	// allowed to be logged.
	// By default, is set to return warning level.
	MinLevel LevelProvider

	// MaxLevel is a function that returns the maximum level
	// allowed to be logged.
	// By default, is set to return error level.
	MaxLevel LevelProvider

	// LevelLabels is a map which defines for each level
	// its string representation.
	// By default, "CRITICAL", "ERROR", "WARN", "INFO", "DEBUG" labels are used.
	LevelLabels map[Level]string

	// LevelKey is the key under which level is found.
	// By default, is set to "lvl".
	LevelKey string

	// TimeKey is the key under which the current log's time is found.
	// By default, is set to "date".
	TimeKey string

	// Time is a function that returns the time to be logged.
	// By default, is set to UTC time formatted as RFC3339Nano.
	Time Provider

	// SourceKey is the key under which caller filename and line are found.
	// It can be set to an empty string if you want to disable this information
	// from logs.
	// By default, is set to "src".
	SourceKey string

	// Source is a provider that returns the source where the log occurred
	// in the call stack.
	Source Provider

	// AdditionalKeyValues holds additional key-values that will be stored
	// with each log.
	// Example: you may want to log your application version or name or
	// environment (dev/stage/production/...), etc.
	// The value can be a Provider for dynamically retrieve a value at runtime.
	AdditionalKeyValues []any

	// ErrHandler callback to process errors that occurred during logging.
	// By design, the logger contract does not return errors from its methods
	// as you most probably use it for this purpose, to log an error, and
	// if an error rises in this process what can you do?
	// You may want to log the error with the standard logger if it
	// suits your needs for example.
	// Source of errors might come from IO errors / formatting errors.
	// By default, is set to a no-op ErrorHandler which disregards the error.
	ErrHandler ErrorHandler
}

// LevelProvider is a function that provides at runtime the min/max
// level allowed to be logged.
type LevelProvider func() Level

// Provider is a function that returns at runtime a value.
type Provider func() any

// ErrorHandler is a callback to handle internal logging errors.
// It accepts as 1st param the internal logging error.
// It accepts as 2nd param the key-values log entry on which error occurred.
type ErrorHandler func(err error, keyValues []any)

// NopErrorHandler is "no-op" error handler for any error occurred
// during log process. It simply ignores the error.
var NopErrorHandler ErrorHandler = func(_ error, _ []any) {}

// NewCommonOpts instantiates a default configured CommonOpts object.
// You can start customization of fields from this object.
func NewCommonOpts() *CommonOpts {
	return &CommonOpts{
		MinLevel: FixedLevelProvider(LevelWarning),
		MaxLevel: FixedLevelProvider(LevelCritical),
		LevelLabels: map[Level]string{
			LevelCritical: "CRITICAL",
			LevelError:    "ERROR",
			LevelWarning:  "WARN",
			LevelInfo:     "INFO",
			LevelDebug:    "DEBUG",
		},
		LevelKey:   defaultOptLevelKey,
		TimeKey:    defaultOptTimeKey,
		Time:       UTCTimeProvider(time.RFC3339Nano),
		SourceKey:  defaultOptSourceKey,
		Source:     SourceProvider(4, 0),
		ErrHandler: NopErrorHandler,
	}
}

// BetweenMinMax returns true if passed level is found in
// [MinLevel, MaxLevel] interval, false otherwise.
func (opts *CommonOpts) BetweenMinMax(lvl Level) bool {
	return lvl >= opts.MinLevel() && lvl <= opts.MaxLevel()
}

// WithDefaultKeyValues returns keyValues enriched with default ones.
func (opts *CommonOpts) WithDefaultKeyValues(lvl Level, keyValues ...any) []any {
	keyVals := make([]any, 0, 6+len(opts.AdditionalKeyValues)+len(keyValues))
	keyValues = AppendNoValue(keyValues)
	keyVals = append(keyVals, opts.TimeKey, opts.Time())
	if lvl != LevelNone {
		keyVals = append(keyVals, opts.LevelKey, opts.LevelLabels[lvl])
	}
	if opts.SourceKey != "" {
		source := opts.Source()
		if source != "" {
			keyVals = append(keyVals, opts.SourceKey, source)
		}
	}

	for i := 0; i < len(opts.AdditionalKeyValues); i += 2 {
		key := opts.AdditionalKeyValues[i]
		value := opts.AdditionalKeyValues[i+1]
		valueProvider, isProvider := value.(Provider)
		if isProvider {
			value = valueProvider()
		}
		keyVals = append(keyVals, key, value)
	}

	keyVals = append(keyVals, keyValues...)

	return keyVals
}

// FixedLevelProvider provides a fixed Level returned at each call.
func FixedLevelProvider(lvl Level) LevelProvider {
	return func() Level { return lvl }
}

// EnvLevelProvider provides a level read from OS's ENV at each call.
// If the level environment key is not found, or value stored in it is
// invalid, default provided level is returned.
// As it is called on each log, you may change during application run
// the underlying env without restarting the app, and new configured
// value will be used in place, if suitable.
func EnvLevelProvider(envLvlKey string, defaultLvl Level, levelLabels map[Level]string) LevelProvider {
	labeledLevels := flipLevelLabels(levelLabels)

	return func() Level {
		envLvl := os.Getenv(envLvlKey)
		lvl, found := labeledLevels[envLvl]
		if found {
			return lvl
		}

		return defaultLvl
	}
}

// UTCTimeProvider is a formatted current UTC time provider.
func UTCTimeProvider(format string) Provider {
	return func() any {
		return time.Now().UTC().Format(format)
	}
}

// LocalTimeProvider is a formatted current local time provider.
func LocalTimeProvider(format string) Provider {
	return func() any {
		return time.Now().Format(format)
	}
}

// SourceProvider is a file and line from call stack
// First param is the number of frames to skip in the call stack.
// Second param is number of directories to skip from file name
// backwards to root dir (0 means full path is returned).
func SourceProvider(skipFrames, skipPath int) Provider {
	return func() any {
		_, file, line, ok := runtime.Caller(skipFrames)
		if ok {
			idx := 0
			if skipPath > 0 {
				for skipPath, i := skipPath, len(file)-1; i >= 0 && skipPath > 0; i-- {
					if file[i] == '/' {
						skipPath--
						idx = i
					}
				}
			}

			return file[idx:] + ":" + strconv.FormatInt(int64(line), 10)
		}

		return ""
	}
}

// AppendNoValue is a safety function which adds a "*NoValue*"
// at the end of keyValues slice in case it is odd.
func AppendNoValue(keyValues []any) []any {
	if len(keyValues)%2 == 1 {
		keyValues = append(keyValues, noValue)
	}

	return keyValues
}

// StackErr return "%+v" fmt representation of an error.
// Can be useful to log stack of errors created with
// "github.com/pkg/errors" / "github.com/actforgood/xerr" packages.
//
// Example of usage:
//
//	logger.Error(xlog.ErrorKey, xlog.StackErr(errWithStack), ...)
func StackErr(err error) string {
	return fmt.Sprintf("%+v", err)
}

// flipLevelLabels flips level labels map.
func flipLevelLabels(levelLabels map[Level]string) map[string]Level {
	flippedLevelLabels := make(map[string]Level, len(levelLabels))
	for lvl, label := range levelLabels {
		flippedLevelLabels[label] = lvl
	}

	return flippedLevelLabels
}
