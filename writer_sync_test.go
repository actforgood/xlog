// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/LICENSE.

package xlog_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/actforgood/xlog"
)

func ExampleNewSyncWriter() {
	// In this example we wrap a bytes.Buffer (Writer)
	// with a SyncWriter in order for Write to be
	// concurrent safe.

	inMemoryWriter := new(bytes.Buffer)
	// writer := inMemoryWriter // you can enable this line instead of SyncWriter to check race conditions.
	writer := xlog.NewSyncWriter(inMemoryWriter)
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	wg := sync.WaitGroup{}

	// perform 5 concurrent writes.
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(threadNo int) {
			defer wg.Done()
			threadNoStr := strconv.FormatInt(int64(threadNo+1), 10)
			_, err := writer.Write([]byte(text + threadNoStr + "\n"))
			if err != nil {
				log.Printf("[%d] write error: %v\n", threadNo, err)
			}
		}(i)
	}
	wg.Wait()

	checkOutput := inMemoryWriter.Bytes()
	fmt.Print(string(checkOutput))

	// Unordered output:
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.1
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.4
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.3
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.5
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.2
}

func TestSyncWriter_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		writer       bytes.Buffer
		subject      = xlog.NewSyncWriter(&writer)
		goroutinesNo = 200
		wg           sync.WaitGroup
		expectedSum  = goroutinesNo * (goroutinesNo + 1) / 2
	)

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
			n, err := subject.Write(data)
			assertEqual(t, len(data), n)
			assertNil(t, err)
		}(i)
	}
	wg.Wait()

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
