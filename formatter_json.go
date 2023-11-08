// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"encoding/json"
	"io"
)

// JSONFormatter serializes key-values in JSON format and writes the
// resulted JSON to the writer.
// It returns error if a serialization/writing problem is encountered.
var JSONFormatter Formatter = func(w io.Writer, keyValues []any) error {
	keyValues = AppendNoValue(keyValues)

	// convert log slice into a map.
	keyValueMap := make(map[string]any, len(keyValues)/2)
	for idx := 0; idx < len(keyValues); idx += 2 {
		keyValueMap[stringify(keyValues[idx])] = valueForJSON(keyValues[idx+1])
	}

	// encode key-value map into JSON.
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(keyValueMap)
}

// valueForJSON applies some customization upon a value.
// Currently an error.Error() is taken instead of error itself.
func valueForJSON(v any) any {
	switch val := v.(type) { // nolint
	case error:
		if val != nil {
			return val.Error()
		}
	}

	return v
}
