// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

// Level of logging.
type Level byte

const (
	// LevelNone represents no level.
	// Is used internally for Log() method.
	LevelNone Level = 0

	// LevelDebug is the level for debug logs.
	LevelDebug Level = 10

	// LevelInfo is the level for info logs.
	LevelInfo Level = 20

	// LevelWarning is the level for warning logs.
	LevelWarning Level = 30

	// LevelError is the level for error logs.
	LevelError Level = 40

	// LevelCritical is the level for critical logs.
	LevelCritical Level = 50
)
