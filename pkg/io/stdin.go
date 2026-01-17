package io

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// ByteReader reads bytes from stdin.
func ByteReader(output chan []byte, pool *sync.Pool, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	r := bufio.NewReaderSize(os.Stdin, 1<<20)

	for {
		buf := pool.Get().([]byte)
		n, err := io.ReadFull(r, buf)

		// no more input, close down
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			pool.Put(buf)
			return nil
		}

		// failed to read from stdin
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		// We expect full blocks
		if n != len(buf) {
			pool.Put(buf)
			return fmt.Errorf("unexpected EOF: expected %d bytes, got %d", len(buf), n)
		}

		output <- buf
	}
}

// Float32Reader reads float32s from stdin.
func Float32Reader(output chan []float32, pool *sync.Pool, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	r := bufio.NewReaderSize(os.Stdin, 1<<20)

	for {
		buf := pool.Get().([]float32)
		err := binary.Read(r, binary.LittleEndian, buf)

		// no more input, close down
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			pool.Put(buf)
			return nil
		}

		// failed to read from stdin
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		output <- buf
	}
}
