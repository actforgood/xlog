// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"io"
	"sync"
)

// AsyncLogger is a Logger which writes logs asynchronously.
// Note: if used in a concurrent context, log writes are concurrent safe if only
// one worker is configured to process the logs. Otherwise, log writes are not
// concurrent safe, unless the writer is concurrent safe. See also NewSyncWriter
// and AsyncLoggerWithWorkersNo on this matter.
type AsyncLogger struct {
	// writer logs will be written to.
	writer io.Writer
	// formatter can be set with AsyncLoggerWithFormatter functional option.
	formatter Formatter
	// internal channel where logs are pushed for processing.
	// its buffer size is 256 by default.
	// can be set with AsyncLoggerWithChannelSize functional option.
	entriesChan chan []interface{}
	// no of workers to start for processing entriesChan.
	workersNo int
	// common options for this logger.
	// can be set with AsyncLoggerWithOptions functional option.
	opts *CommonOpts
	// closed flag, true means Close() has been called, from this point forward,
	// entriesChan is waited to be drained, and no further logs are accepted for
	// processing.
	closed bool
	// concurrency semaphore to protect closed flag access.
	closeMu sync.RWMutex
	// wait group to synchronize internal started goroutine(s) with Close method,
	// to wait for entriesChan to be drained, and all logs processed.
	wg sync.WaitGroup
}

// NewAsyncLogger instantiates a new logger object that writes logs
// asynchronously.
// First param is a Writer where logs are written to.
// Example: os.Stdout, a custom opened os.File, an in memory strings.Buffer, etc.
// Second param is/are function option(s) through which you can customize
// the logger. Check for AsyncLoggerWith* options.
func NewAsyncLogger(w io.Writer, opts ...AsyncLoggerOption) *AsyncLogger {
	// instantiate object with default properties.
	logger := &AsyncLogger{
		writer:    w,
		formatter: JSONFormatter,
		workersNo: 1,
	}

	// apply options, if any.
	for _, opt := range opts {
		opt(logger)
	}

	if logger.opts == nil {
		logger.opts = NewCommonOpts()
	}
	// if no option was provided for entriesChan, use default.
	if logger.entriesChan == nil {
		const defaultEntriesChanSize = 256
		logger.entriesChan = make(chan []interface{}, defaultEntriesChanSize)
	}

	// start internal goroutine(s) that will log entries async.
	logger.startWorkers()

	return logger
}

// startWorkers start configured no of goroutines that process logs.
func (logger *AsyncLogger) startWorkers() {
	logger.wg.Add(logger.workersNo)
	worker := logger.logAsync
	for i := 0; i < logger.workersNo; i++ {
		go worker()
	}
}

// logAsync processes logs channel and performs the actual logging.
// it is meant to be called in another goroutine.
func (logger *AsyncLogger) logAsync() {
	defer logger.wg.Done() // notify waiting thread work is finished.

	for keyVals := range logger.entriesChan {
		// format the log.
		if err := logger.formatter(logger.writer, keyVals); err != nil {
			logger.opts.ErrHandler(err, keyVals)
		}
	}
}

// Critical logs application component unavailable, fatal events.
func (logger *AsyncLogger) Critical(keyValues ...interface{}) {
	logger.pushLog(LevelCritical, keyValues...)
}

// Error logs runtime errors that
// should typically be logged and monitored.
func (logger *AsyncLogger) Error(keyValues ...interface{}) {
	logger.pushLog(LevelError, keyValues...)
}

// Warn logs exceptional occurrences that are not errors.
// Example: Use of deprecated APIs, poor use of an API, undesirable things
// that are not necessarily wrong.
func (logger *AsyncLogger) Warn(keyValues ...interface{}) {
	logger.pushLog(LevelWarning, keyValues...)
}

// Info logs interesting events.
// Example: User logs in, SQL logs.
func (logger *AsyncLogger) Info(keyValues ...interface{}) {
	logger.pushLog(LevelInfo, keyValues...)
}

// Debug logs detailed debug information.
func (logger *AsyncLogger) Debug(keyValues ...interface{}) {
	logger.pushLog(LevelDebug, keyValues...)
}

// Log logs arbitrary data.
func (logger *AsyncLogger) Log(keyValues ...interface{}) {
	logger.pushLog(LevelNone, keyValues...)
}

// Close nicely closes logger.
// You should call it to make sure all logs have been processed
// (for example at your application shutdown).
// Once called, any further call to any of the logging methods will be ignored.
func (logger *AsyncLogger) Close() error {
	logger.closeMu.Lock()
	defer logger.closeMu.Unlock()

	if !logger.closed {
		logger.closed = true      // mark logger as closed.
		close(logger.entriesChan) // close log entries chan.
		logger.wg.Wait()          // wait for workers to process any entry left in chan.

		if bw, ok := logger.writer.(*BufferedWriter); ok {
			bw.Stop()
		}
	}

	return nil
}

// isClosed returns true if Close method was called, false otherwise.
func (logger *AsyncLogger) isClosed() bool {
	logger.closeMu.RLock()
	defer logger.closeMu.RUnlock()

	return logger.closed
}

// pushLog sends log entry to internal logs channel.
// Note: this call blocks if internal logs channel is full
// (the rate of producing messages is much higher than consuming one).
// Using AsyncLoggerWithChannelSize to set a higher value to increase
// throughput in such case can be helpful. Also setting more workers can
// be helpful, see AsyncLoggerWithWorkersNo.
func (logger *AsyncLogger) pushLog(lvl Level, keyValues ...interface{}) {
	// ignore log conditions check.
	if !logger.opts.BetweenMinMax(lvl) {
		return
	}

	// enrich passed key values with default ones.
	keyVals := logger.opts.WithDefaultKeyValues(lvl, keyValues...)

	// send log for async processing.
	if !logger.isClosed() {
		logger.entriesChan <- keyVals
	}
}
