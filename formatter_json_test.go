// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/actforgood/xlog"
)

func TestJSONFormatter_successfullyWritesJSON(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.JSONFormatter
		dummy     = dummyStringer{Name: "John Doe"}
		someErr   = errors.New("test err.Error() is serialized")
		keyValues = []any{
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
	writtenBytes := writer.Bytes()
	var kvMap map[string]any
	if err := json.Unmarshal(writtenBytes, &kvMap); err != nil {
		t.Fatal(err.Error())
	}
	assertEqual(t, 7, len(kvMap))
	assertEqual(t, "bar", kvMap["foo"])
	assertEqual(t, 34, int(kvMap["age"].(float64)))
	assertEqual(t, 123.456, kvMap["computation"])
	assertEqual(t, "ten", kvMap["10"])
	assertEqual(t, 3, len(kvMap["ints-slice"].([]any)))
	assertEqual(
		t,
		map[string]any{"Name": "John Doe"},
		kvMap["dummyStringer: John Doe"],
	)
	assertEqual(t, someErr.Error(), kvMap["err"])
}

func TestJSONFormatter_returnsWriteErr(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xlog.JSONFormatter
		dummy     = dummyStringer{Name: "John Doe"}
		keyValues = []any{
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

func BenchmarkJSONFormatter(b *testing.B) {
	var (
		subject = xlog.JSONFormatter
		dummy   = dummyStringer{Name: "John Doe"}
		input   = []any{
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
