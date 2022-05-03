// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

import (
	"bytes"
	"io"
	"sync"

	"github.com/go-logfmt/logfmt"
)

type logfmtEncoder struct {
	*logfmt.Encoder
	buf bytes.Buffer
}

// Reset resets the encoder and its buffer.
func (enc *logfmtEncoder) Reset() {
	enc.Encoder.Reset()
	enc.buf.Reset()
}

// Encode encodes given key values in logfmt format.
func (enc *logfmtEncoder) Encode(keyValues ...interface{}) error {
	if err := enc.EncodeKeyvals(keyValues...); err != nil {
		return err
	}

	return enc.EndRecord()
}

var logfmtEncoderPool = sync.Pool{
	New: func() interface{} {
		enc := new(logfmtEncoder)
		enc.Encoder = logfmt.NewEncoder(&enc.buf)

		return enc
	},
}

// LogfmtFormatter serializes key-values in logfmt format and writes the
// resulted bytes to the writer.
// It returns error if a serialization/writing problem is encountered.
// More about logfmt can be found here: https://brandur.org/logfmt .
var LogfmtFormatter Formatter = func(w io.Writer, keyValues []interface{}) error {
	keyValues = AppendNoValue(keyValues)

	enc := logfmtEncoderPool.Get().(*logfmtEncoder)
	enc.Reset()
	defer logfmtEncoderPool.Put(enc)

	if err := enc.Encode(keyValues...); err != nil {
		return err
	}

	if _, err := w.Write(enc.buf.Bytes()); err != nil {
		return err
	}

	return nil
}
