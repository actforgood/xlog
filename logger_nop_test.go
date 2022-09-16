// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"testing"

	"github.com/actforgood/xlog"
)

func TestNopLogger_doesNothing(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject xlog.Logger = xlog.NopLogger{} // check also satisfies contract
		kv                  = getInputKeyValues()
	)

	// act
	subject.Log(kv)
	subject.Debug(kv)
	subject.Info(kv)
	subject.Warn(kv)
	subject.Error(kv)
	subject.Critical(kv)
	err := subject.Close()
	assertNil(t, err)
}
