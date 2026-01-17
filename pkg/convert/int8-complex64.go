package convert

import "sync"

const scale = float32(1.0 / 128.0)

func Int8ToComplex64(input chan []byte, inPool *sync.Pool, output chan []complex64, outPool *sync.Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	for inBuf := range input {
		outBuf := outPool.Get().([]complex64)

		// Convert block
		for i := 0; i < len(inBuf)/2; i++ {
			I := float32(int8(inBuf[2*i])) * scale
			Q := float32(int8(inBuf[2*i+1])) * scale
			outBuf[i] = complex(I, Q)
		}

		inPool.Put(inBuf)
		output <- outBuf
	}
}
