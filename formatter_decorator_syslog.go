//go:build !windows && !nacl && !plan9
// +build !windows,!nacl,!plan9

// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"bytes"
	"errors"
	"io"
	"log/syslog"
)

// syslogWriter is an interface wrapping stdlib syslog Writer.
type syslogWriter interface {
	// Write sends a log message to the syslog daemon.
	Write([]byte) (int, error)
	// Close closes a connection to the syslog daemon.
	Close() error
	// Emerg logs a message with severity LOG_EMERG, ignoring the severity
	// passed to New.
	Emerg(string) error
	// Alert logs a message with severity LOG_ALERT, ignoring the severity
	// passed to New.
	Alert(string) error
	// Crit logs a message with severity LOG_CRIT, ignoring the severity
	// passed to New.
	Crit(string) error
	// Err logs a message with severity LOG_ERR, ignoring the severity
	// passed to New.
	Err(string) error
	// Warning logs a message with severity LOG_WARNING, ignoring the
	// severity passed to New.
	Warning(string) error
	// Notice logs a message with severity LOG_NOTICE, ignoring the
	// severity passed to New.
	Notice(string) error
	// Info logs a message with severity LOG_INFO, ignoring the severity
	// passed to New.
	Info(string) error
	// Debug logs a message with severity LOG_DEBUG, ignoring the severity
	// passed to New.
	Debug(string) error
}

// SyslogPrefixCee defines the @cee prefix for structured logging in syslog.
// See: http://cee.mitre.org/language/1.0-beta1/clt.html#appendix-1-cee-over-syslog-transport-mapping .
const SyslogPrefixCee = "@cee:"

// ErrNotSyslogWriter is the error returned in case the writer is not syslog specific.
var ErrNotSyslogWriter = errors.New("the writer should be a *syslog.Writer")

// SyslogLevelProvider is a function that extracts the syslog level.
type SyslogLevelProvider func(keyValues []interface{}) syslog.Priority

const noLevel = syslog.Priority(-100000)

// NewDefaultSyslogLevelProvider returns a SyslogLevelProvider that maps xlog default Levels
// to their appropriate syslog Levels.
func NewDefaultSyslogLevelProvider(opts *CommonOpts) SyslogLevelProvider {
	levelsMap := make(map[interface{}]syslog.Priority, 5)
	for lvl, label := range opts.LevelLabels {
		switch lvl {
		case LevelDebug:
			levelsMap[label] = syslog.LOG_DEBUG
		case LevelInfo:
			levelsMap[label] = syslog.LOG_INFO
		case LevelWarning:
			levelsMap[label] = syslog.LOG_WARNING
		case LevelError:
			levelsMap[label] = syslog.LOG_ERR
		case LevelCritical:
			levelsMap[label] = syslog.LOG_CRIT
		}
	}

	return NewExtractFromKeySyslogLevelProvider(opts.LevelKey, levelsMap)
}

// NewExtractFromKeySyslogLevelProvider extracts the value of given key as first param
// and returns the syslog level from provided map as second param.
func NewExtractFromKeySyslogLevelProvider(
	key string,
	syslogLevels map[interface{}]syslog.Priority,
) SyslogLevelProvider {
	return func(keyValues []interface{}) syslog.Priority {
		syslogLevel, found := syslogLevels[extractKeyValue(key, keyValues)]
		if found {
			return syslogLevel
		}

		return noLevel
	}
}

// SyslogFormatter is a decorator which writes another formatter 's output to system syslog.
// The second param is a function that knows to return a syslog level for the current log.
// You can use [NewDefaultSyslogLevelProvider] / [NewExtractFromKeySyslogLevelProvider] or custom provider
// (maybe you want to support other syslog levels - for example nothing stops you from doing this:
// logger.Log("lvl","NOTICE", ...) and map also "NOTICE" to [syslog.LOG_NOTICE]).
// The third param is a prefix to be written with each log. You'll pass here empty string or [SyslogPrefixCee].
var SyslogFormatter = func(
	formatter Formatter,
	syslogLevelProvider SyslogLevelProvider,
	prefix string,
) Formatter {
	return func(w io.Writer, keyValues []interface{}) error {
		sw, ok := w.(syslogWriter)
		if !ok {
			return ErrNotSyslogWriter
		}
		keyValues = AppendNoValue(keyValues)

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)

		if prefix != "" {
			_, _ = buf.WriteString(prefix)
		}

		if err := formatter(buf, keyValues); err != nil {
			return err
		}

		syslogLevel := syslogLevelProvider(keyValues)
		switch syslogLevel {
		case syslog.LOG_EMERG:
			return sw.Emerg(buf.String())
		case syslog.LOG_ALERT:
			return sw.Alert(buf.String())
		case syslog.LOG_CRIT:
			return sw.Crit(buf.String())
		case syslog.LOG_ERR:
			return sw.Err(buf.String())
		case syslog.LOG_WARNING:
			return sw.Warning(buf.String())
		case syslog.LOG_NOTICE:
			return sw.Notice(buf.String())
		case syslog.LOG_INFO:
			return sw.Info(buf.String())
		case syslog.LOG_DEBUG:
			return sw.Debug(buf.String())
		default:
			_, err := sw.Write(buf.Bytes())

			return err
		}
	}
}
