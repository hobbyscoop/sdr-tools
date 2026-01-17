package io

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

// Complex64Writer writes complex samples to stdout.
// If wg is not nil, it will be marked as done when the function returns.
func Complex64Writer(input chan []complex64, pool *sync.Pool, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()
	}

	w := bufio.NewWriterSize(os.Stdout, 1<<20)
	defer w.Flush()

	for buf := range input {
		err := binary.Write(w, binary.LittleEndian, buf)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
		pool.Put(buf)
	}
	return nil
}

// Int16Writer writes int16 samples to stdout.
// If wg is not nil, it will be marked as done when the function returns.
func Int16Writer(input chan []int16, pool *sync.Pool, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()
	}

	w := bufio.NewWriterSize(os.Stdout, 1<<20)
	defer w.Flush()

	for buf := range input {
		err := binary.Write(w, binary.LittleEndian, buf)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
		pool.Put(buf)
	}
	return nil
}
