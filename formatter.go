// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"fmt"
	"io"
	"strconv"
)

// Formatter writes the provided key-values in a given format.
// Returns error in case something goes wrong.
type Formatter func(w io.Writer, keyValues []any) error

// stringify returns string representation of an interface.
func stringify(i any) string {
	switch data := i.(type) {
	case string:
		return data
	case fmt.Stringer:
		return data.String()
	case int:
		return strconv.FormatInt(int64(data), 10)
	}

	return fmt.Sprint(i)
}
