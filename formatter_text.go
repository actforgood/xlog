// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"bytes"
	"io"
)

// TextFormatter provides a more human friendly custom format.
// This formatter does not comply with any kind of well known standard.
// It can be used for example for local dev environment.
// Example of output: "TIME SOURCE LEVEL MESSAGE KEY1=VALUE1 KEY2=VALUE2 ...".
var TextFormatter = func(opts *CommonOpts) Formatter {
	return func(w io.Writer, keyValues []any) error {
		keyValues = AppendNoValue(keyValues)

		var (
			time, level, source, msg  string
			finalOutBuf, extraInfoBuf bytes.Buffer
			key, value                any
		)
		finalOutBuf.Grow(64)
		extraInfoBuf.Grow(64)

		for idx := 0; idx < len(keyValues); idx += 2 {
			key = keyValues[idx]
			value = keyValues[idx+1]
			switch key {
			case opts.LevelKey:
				level = stringify(value)
			case opts.TimeKey:
				time = stringify(value)
			case opts.SourceKey:
				source = stringify(value)
			case MessageKey:
				msg = stringify(value)
			default:
				_, _ = extraInfoBuf.WriteString(stringify(key))
				_ = extraInfoBuf.WriteByte('=')
				_, _ = extraInfoBuf.WriteString(stringify(value))
				_ = extraInfoBuf.WriteByte(' ')
			}
		}

		appendTextFinalOutput(&finalOutBuf, []byte(time))
		appendTextFinalOutput(&finalOutBuf, []byte(source))
		appendTextFinalOutput(&finalOutBuf, []byte(level))
		appendTextFinalOutput(&finalOutBuf, []byte(msg))
		finalOut := append(finalOutBuf.Bytes(), extraInfoBuf.Bytes()...)
		finalOut[len(finalOut)-1] = '\n' // replace last space with new line

		_, err := w.Write(finalOut)

		return err
	}
}

func appendTextFinalOutput(buf *bytes.Buffer, info []byte) {
	if len(info) > 0 {
		_, _ = buf.Write(info)
		_ = buf.WriteByte(' ')
	}
}
