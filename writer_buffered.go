// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog

import (
	"bufio"
	"io"
	"sync"
	"time"
)

// default buffer size (4 Kb).
// It complies with the default chosen buffer size by Go from here:
// https://github.com/golang/go/blob/go1.17.3/src/bufio/bufio.go#L19 .
const defaultBufSize = 1024 * 4

// default interval logs will be flushed regardless buffer written bytes.
const defaultFlushInterval = 10 * time.Second

// BufferedWriter decorates an io.Writer so that written bytes are buffered.
// It is concurrent safe to use.
// It has the capability of auto-flushing the buffer, time interval based.
type BufferedWriter struct {
	// original writer data is written to.
	origWriter io.Writer
	// the buffer size (minimum amount of bytes that will trigger one Write).
	bufSize int
	// bufWriter is the buffered writer decorator.
	bufWriter *bufio.Writer
	// the duration to Flush so far collected bytes, regardless
	// if buffer contains something / is full or not.
	flushInterval time.Duration
	// ticker is used to trigger Flush so far collected bytes
	// regardless if buffer is full or not.
	ticker *time.Ticker
	// the channel to sync with internal ticking goroutine when
	// buffer is stopped.
	stopFlushCh chan struct{}
	// if flag is true means Stop() has been called, from this point forward,
	// no further writes are accepted for and flush goroutine stops.
	stopped bool
	// concurrency semaphore to protect stopped flag access.
	stopMu sync.RWMutex
	// wait group to synchronize internal started goroutine(s) with Store method,
	// to wait for any left byte to be written to original writer.
	wg sync.WaitGroup
	// concurrency semaphore to protect access to buffWriter's operations.
	mu sync.Mutex
}

// NewBufferedWriter instantiates a new buffered writer.
func NewBufferedWriter(w io.Writer, opts ...BufferedWriterOption) *BufferedWriter {
	// instantiate object with default properties.
	bufferedWriter := &BufferedWriter{
		origWriter:    w,
		bufSize:       defaultBufSize,
		flushInterval: defaultFlushInterval,
		stopFlushCh:   make(chan struct{}, 1),
	}

	// apply options, if any.
	for _, opt := range opts {
		opt(bufferedWriter)
	}
	bufferedWriter.bufWriter = bufio.NewWriterSize(
		bufferedWriter.origWriter,
		bufferedWriter.bufSize,
	)

	// start auto-flushing goroutine, if enabled.
	if bufferedWriter.flushInterval > 0 {
		bufferedWriter.ticker = time.NewTicker(bufferedWriter.flushInterval)
		bufferedWriter.wg.Add(1)
		go bufferedWriter.flushAsync()
	}

	return bufferedWriter
}

// Write writes given bytes to the decorated writer (buffered).
// Returns no. of bytes written, or an error.
func (bw *BufferedWriter) Write(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if !bw.isStopped() {
		n, err := bw.bufWriter.Write(p)
		if err != nil {
			// reset to clear the error, otherwise will be returned at any future write.
			bw.bufWriter.Reset(bw.origWriter)
		}

		return n, err
	}

	return 0, nil
}

// flushAsync periodically flushes the buffer.
func (bw *BufferedWriter) flushAsync() {
	defer func() {
		bw.ticker.Stop() // stop the ticker to avoid mem leaks.
		bw.wg.Done()     // notify waiting thread work is finished.
	}()

	// for is executing infinitely,
	// waiting for interval to elapse, or for a stop signal.
	for {
		select {
		case <-bw.ticker.C:
			bw.flush()
		case <-bw.stopFlushCh:
			return
		}
	}
}

// flush simply flushes the buffered writer,
// writing all (if any) stored bytes.
func (bw *BufferedWriter) flush() {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if err := bw.bufWriter.Flush(); err != nil {
		// reset to clear the error, otherwise will be returned at any future write.
		bw.bufWriter.Reset(bw.origWriter)
	}
}

// Stop marks the writer as stopped.
// You should call it to make sure all data have been processed.
// Once called any further Write will be ignored.
func (bw *BufferedWriter) Stop() {
	bw.stopMu.Lock()
	defer bw.stopMu.Unlock()

	if !bw.stopped {
		bw.stopped = true     // mark writer as stopped.
		close(bw.stopFlushCh) // signal flush goroutine to stop by closing the chan.
		bw.wg.Wait()          // wait for flush goroutine to finish.
		bw.flush()            // trigger a flush to store any buffered data.
	}
}

// isStopped returns true if Stop method was called, false otherwise.
func (bw *BufferedWriter) isStopped() bool {
	bw.stopMu.RLock()
	defer bw.stopMu.RUnlock()

	return bw.stopped
}

// BufferedWriterOption defines optional function for configuring
// a buffered writer.
type BufferedWriterOption func(*BufferedWriter)

// BufferedWriterWithSize sets desired buffer size.
// 4Kb is used by default.
func BufferedWriterWithSize(bufSize int) BufferedWriterOption {
	return func(bw *BufferedWriter) {
		bw.bufSize = bufSize
	}
}

// BufferedWriterWithFlushInterval sets desired auto-flush interval.
// 10s is used by default.
// Pass a value <=0 if you want to disable the interval based auto-flush.
func BufferedWriterWithFlushInterval(flushInterval time.Duration) BufferedWriterOption {
	return func(bw *BufferedWriter) {
		bw.flushInterval = flushInterval
	}
}
