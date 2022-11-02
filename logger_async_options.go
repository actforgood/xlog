// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

// AsyncLoggerOption defines optional function for configuring
// an async logger.
type AsyncLoggerOption func(*AsyncLogger)

// AsyncLoggerWithChannelSize sets internal buffered channel's size,
// through which logs are processed.
// If not called, defaults to a 256 size.
// Note: logging blocks if internal logs channel is full
// (the rate of producing messages is much higher than consuming one).
// Using AsyncLoggerWithChannelSize to set a higher value to increase
// throughput in such case can be helpful.
func AsyncLoggerWithChannelSize(logsChanSize uint16) AsyncLoggerOption {
	return func(logger *AsyncLogger) {
		logger.entriesChan = make(chan []interface{}, logsChanSize)
	}
}

// AsyncLoggerWithWorkersNo sets the no. of workers to process
// internal logs channel.
// If not called, defaults to 1.
// Note: logging blocks if internal logs channel is full
// (the rate of producing messages is more high than consuming one).
// Using AsyncLoggerWithWorkersNo to increase number of workers
// might be helpful.
// Note: having a value greater than 1 implies that logs may not
// necessarily be written in their timestamp order. Example: goroutine #1
// reads a log entry with timestamp sec :00.999 and goroutine #2 reads
// a log entry with timestamp sec :01.000 in the same time, but goroutine #2
// is the first one to write its log entry.
// Note: having a value greater than 1 implies that the underlying writer
// is concurrent safe for Writes.
func AsyncLoggerWithWorkersNo(workersNo uint16) AsyncLoggerOption {
	return func(logger *AsyncLogger) {
		logger.workersNo = int(workersNo)
	}
}

// AsyncLoggerWithFormatter sets desired formatter for the logs.
// The JSON formatter is used by default.
func AsyncLoggerWithFormatter(formatter Formatter) AsyncLoggerOption {
	return func(logger *AsyncLogger) {
		logger.formatter = formatter
	}
}

// AsyncLoggerWithOptions sets the common options.
// A [NewCommonOpts] is used by default.
func AsyncLoggerWithOptions(opts *CommonOpts) AsyncLoggerOption {
	return func(logger *AsyncLogger) {
		logger.opts = opts
	}
}
