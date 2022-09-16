// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog_test

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/actforgood/xlog"
)

func TestBufferedWriter_Write_Stop_isReallyBuffered(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  = new(MockWriter)
		subject = xlog.NewBufferedWriter(
			writer,
			xlog.BufferedWriterWithSize(2),          // we set size to 2 bytes, and we'll write 1 byte
			xlog.BufferedWriterWithFlushInterval(0), // disable auto-flushing
		)
		dummyByte byte = '\n'
	)
	defer subject.Stop()
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		assertEqual(t, []byte{dummyByte}, p)

		return len(p), nil
	})

	// act - write a dummy byte.
	n, err := subject.Write([]byte{dummyByte})

	// assert - we check no byte has been written to underlying writer.
	assertEqual(t, 1, n)
	assertNil(t, err)
	assertEqual(t, 0, writer.WriteCallsCount())

	// act - trigger flushing.
	subject.Stop()

	// assert - check dummy byte was written.
	assertEqual(t, 1, writer.WriteCallsCount())
}

func TestBufferedWriter_Write_Stop_autoFlushWorks(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  = new(MockWriter)
		subject = xlog.NewBufferedWriter(
			writer,
			xlog.BufferedWriterWithSize(2), // we set size to 2 bytes, and we'll write 1 byte
			xlog.BufferedWriterWithFlushInterval(700*time.Millisecond), // enable auto-flushing at 0.7 sec interval
		)
		dummyByte byte = '\n'
	)
	defer subject.Stop()
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		assertEqual(t, []byte{dummyByte}, p)

		return len(p), nil
	})

	// act - write a dummy byte.
	n, err := subject.Write([]byte{dummyByte})

	// assert - we check no byte has been written to underlying writer and writer.Write did not get called.
	assertEqual(t, 1, n)
	assertNil(t, err)
	assertEqual(t, 0, writer.WriteCallsCount())

	// wait 1s - within this time auto-flushing should have been called.
	time.Sleep(1 * time.Second)

	// assert - check dummy byte was written.
	assertEqual(t, 1, writer.WriteCallsCount())
}

func TestBufferedWriter_Stop_nothingGetsWrittenAfterStop(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  = new(MockWriter)
		subject = xlog.NewBufferedWriter(
			writer,
			xlog.BufferedWriterWithSize(1),          // we set size to 1 byte.
			xlog.BufferedWriterWithFlushInterval(0), // disable auto-flushing
		)
		dummyByte byte = '\n'
	)
	defer subject.Stop()
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		assertEqual(t, []byte{dummyByte, dummyByte}, p)

		return len(p), nil
	})

	// act - write 2 dummy bytes, 1 flush will be triggered.
	n, err := subject.Write([]byte{dummyByte, dummyByte})

	// assert - we check the bytes were written on underlying writer.
	assertEqual(t, 2, n)
	assertNil(t, err)
	assertEqual(t, 1, writer.WriteCallsCount())

	// act - stop & write again
	subject.Stop()
	n, err = subject.Write([]byte{dummyByte, dummyByte, dummyByte})

	// assert - check nothing was written.
	assertEqual(t, 0, n)
	assertNil(t, err)
	assertEqual(t, 1, writer.WriteCallsCount()) // calls count is still 1
}

func TestBufferedWriter_Write_writeErrorGetsReset(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  = new(MockWriter)
		subject = xlog.NewBufferedWriter(
			writer,
			xlog.BufferedWriterWithSize(1),          // we set size to 1 byte.
			xlog.BufferedWriterWithFlushInterval(0), // disable auto-flushing.
		)
		dummyByte byte = '\n'
	)
	defer subject.Stop()
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		if writer.WriteCallsCount() == 1 {
			assertEqual(t, []byte{dummyByte, dummyByte}, p)

			return 0, ErrWrite
		}
		assertEqual(t, []byte{dummyByte, dummyByte, dummyByte}, p)

		return len(p), nil
	})

	// act - write 2 dummy bytes.
	n, err := subject.Write([]byte{dummyByte, dummyByte})

	// assert
	assertEqual(t, 0, n)
	assertTrue(t, errors.Is(err, ErrWrite))

	// act - write 3 dummy bytes, successfully this time.
	n, err = subject.Write([]byte{dummyByte, dummyByte, dummyByte})

	// assert
	assertEqual(t, 3, n)
	assertNil(t, err)

	assertEqual(t, 2, writer.WriteCallsCount())
}

func TestBufferedWriter_Write_autoFlushErrorGetsReset(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  = new(MockWriter)
		subject = xlog.NewBufferedWriter(
			writer,
			xlog.BufferedWriterWithSize(1), // we set size to 1 byte.
			xlog.BufferedWriterWithFlushInterval(700*time.Millisecond),
		)
		dummyByte byte = '\n'
	)
	defer subject.Stop()
	writer.SetWriteCallback(func(p []byte) (n int, err error) {
		if writer.WriteCallsCount() == 1 { // auto Flush()
			assertEqual(t, []byte{dummyByte}, p)

			return 0, ErrWrite
		}
		assertEqual(t, []byte{dummyByte, dummyByte, dummyByte}, p)

		return len(p), nil
	})

	// act - write 1 dummy byte.
	n, err := subject.Write([]byte{dummyByte})

	// assert
	assertEqual(t, 1, n)
	assertNil(t, err)

	// wait 1s - within this time auto-flushing should have been called.
	time.Sleep(1 * time.Second)

	// act - write 3 dummy bytes, successfully this time.
	n, err = subject.Write([]byte{dummyByte, dummyByte, dummyByte})

	// assert
	assertEqual(t, 3, n)
	assertNil(t, err)

	assertEqual(t, 2, writer.WriteCallsCount())
}

func TestBufferedWriter_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer  bytes.Buffer
		subject = xlog.NewBufferedWriter(
			&writer,
			xlog.BufferedWriterWithSize(2),
			xlog.BufferedWriterWithFlushInterval(50*time.Millisecond),
		)
		goroutinesNo = 200
		wg           sync.WaitGroup
		expectedSum  = goroutinesNo * (goroutinesNo + 1) / 2
	)
	defer subject.Stop()

	// act
	// N threads will write their own number.
	// At the end we will check the sum of all written numbers
	// (we know from math that sum of first N consecutive numbers is N * (N+1) / 2).
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(threadNo int) {
			defer wg.Done()
			data := []byte(strconv.FormatInt(int64(threadNo+1), 10))
			data = append(data, '\n')
			_, err := subject.Write(data)
			assertNil(t, err)
		}(i)
	}
	wg.Wait()
	subject.Stop()

	// assert
	linesCount := 0
	sum := 0
	for {
		line, err := writer.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Error(err.Error())

			continue
		}
		linesCount++
		line = line[0 : len(line)-1] // remove ending '\n'

		data, err := strconv.Atoi(string(line))
		if err != nil {
			t.Error(err.Error())
		}
		sum += data
	}
	assertEqual(t, goroutinesNo, linesCount)
	assertEqual(t, expectedSum, sum)
}
