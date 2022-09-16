// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"bytes"
	"io"
	"sync"

	"github.com/getsentry/sentry-go"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// extractLevel searches for level label and returns its byte representation.
func extractLevel(labeledLevels map[string]Level, levelKey string, keyValues []interface{}) Level {
	if lvl, found := labeledLevels[extractKeyValue(levelKey, keyValues).(string)]; found {
		return lvl
	}

	return LevelNone
}

// extractKeyValue searches for a key and returns its value.
func extractKeyValue(key string, keyValues []interface{}) interface{} {
	for idx := 0; idx < len(keyValues); idx += 2 {
		if keyValues[idx] == key && idx+1 < len(keyValues) {
			return keyValues[idx+1]
		}
	}

	return ""
}

// SentryFormatter is a decorator which sends another formatter 's output to Sentry.
// The writer from the Logger should be io.Discard, as it uses internally a bytes.Buffer.
var SentryFormatter = func(formatter Formatter, hub *sentry.Hub, opts *CommonOpts) Formatter {
	var (
		mu             sync.Mutex
		sentryLevelMap = map[Level]sentry.Level{
			LevelDebug:    sentry.LevelDebug,
			LevelInfo:     sentry.LevelInfo,
			LevelWarning:  sentry.LevelWarning,
			LevelError:    sentry.LevelError,
			LevelCritical: sentry.LevelFatal,
			LevelNone:     sentry.Level(""),
		}
		labeledLevels = flipLevelLabels(opts.LevelLabels)
	)

	return func(_ io.Writer, keyValues []interface{}) error {
		keyValues = AppendNoValue(keyValues)

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)

		if err := formatter(buf, keyValues); err != nil {
			return err
		}

		sentryLevel := sentryLevelMap[extractLevel(labeledLevels, opts.LevelKey, keyValues)]

		mu.Lock()
		defer mu.Unlock()

		hub.Scope().SetLevel(sentryLevel)
		_ = hub.CaptureMessage(buf.String())

		return nil
	}
}
