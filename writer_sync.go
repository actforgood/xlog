// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import (
	"io"
	"sync"
)

// syncWriter decorates an io.Writer so that each call to Write is synchronized
// with a mutex, making is safe for concurrent use by multiple goroutines.
// It should be used if writer's Write method is not thread safe.
// For example an os.File is safe, so it doesn't need this wrapper,
// on the other hand, a bytes.Buffer is not.
type syncWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

// NewSyncWriter instantiates a new Writer decorated
// with a mutex, making is safe for concurrent use by multiple goroutines.
func NewSyncWriter(w io.Writer) io.Writer {
	return &syncWriter{
		w:  w,
		mu: new(sync.Mutex),
	}
}

// Write writes given bytes to the decorated writer.
// Returns no. of bytes written, or an error.
func (sw syncWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	return sw.w.Write(p)
}
