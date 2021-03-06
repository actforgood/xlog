// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog_test

import (
	"io/ioutil"
	"os"

	"github.com/actforgood/xlog"
)

// Note: this file contains some common utilities used in tests.

// staticTime is a static time returned as Source for logs.
const staticTime = "2021-11-30T16:01:20Z"

// staticTimeProvider returns a static time for logs.
var staticTimeProvider = func() interface{} {
	return staticTime
}

// callMethodByLevel calls appropriate method on subject based on provided level.
func callMethodByLevel(subject xlog.Logger, lvl xlog.Level) {
	switch lvl {
	case xlog.LevelNone:
		subject.Log(getInputKeyValues()...)
	case xlog.LevelDebug:
		subject.Debug(getInputKeyValues()...)
	case xlog.LevelInfo:
		subject.Info(getInputKeyValues()...)
	case xlog.LevelWarning:
		subject.Warn(getInputKeyValues()...)
	case xlog.LevelError:
		subject.Error(getInputKeyValues()...)
	case xlog.LevelCritical:
		subject.Critical(getInputKeyValues()...)
	}
}

// getExpectedKeyValues returns final key values to be logged.
func getExpectedKeyValues(lvl xlog.Level, labels map[xlog.Level]string) []interface{} {
	inputKeyValues := getInputKeyValues()
	additionalKeyValues := getAdditionalKeyValues()

	expectedKeyValues := make([]interface{}, 0, 6+len(inputKeyValues)+len(additionalKeyValues))
	expectedKeyValues = append(expectedKeyValues, "date", staticTime)
	if lvl != xlog.LevelNone {
		expectedKeyValues = append(expectedKeyValues, "lvl", labels[lvl])
	}
	expectedKeyValues = append(expectedKeyValues, additionalKeyValues...)
	expectedKeyValues = append(expectedKeyValues, inputKeyValues...)

	return expectedKeyValues
}

// getInputKeyValues returns input test data to be logged.
func getInputKeyValues() []interface{} {
	return []interface{}{
		"foo", "bar",
		"no", 10,
	}
}

// getAdditionalKeyValues returns additional key values to be logged.
func getAdditionalKeyValues() []interface{} {
	return []interface{}{
		"extraKey", "extraValue",
	}
}

// setUpFile creates a new file for writing logs in it on the disk.
func setUpFile(testName string) *os.File {
	f, err := ioutil.TempFile("", testName+".log-")
	if err != nil {
		panic(err)
	}

	return f
}

// tearDownFile closes test log file and deletes it from the disk.
func tearDownFile(f *os.File) {
	_ = f.Close()
	_ = os.Remove(f.Name())
}

// getBenchmarkKeyVals returns some key vals used in benchmark tests.
func getBenchmarkKeyVals() []interface{} {
	dummy := dummyStringer{Name: "John"}

	return []interface{}{"test", 123, "aaa", "bbb", "dummy", dummy}
}

// dummyStringer is dummy Stringer used in tests.
type dummyStringer struct {
	Name string
}

func (ds dummyStringer) String() string {
	return "dummyStringer: " + ds.Name
}
