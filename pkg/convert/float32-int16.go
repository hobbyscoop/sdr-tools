package convert

import "sync"

func Float32ToInt16(input chan []float32, inPool *sync.Pool, output chan []int16, outPool *sync.Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	for inBuf := range input {
		outBuf := outPool.Get().([]int16)

		// Convert block
		for i := 0; i < len(inBuf); i++ {
			outBuf[i] = int16(inBuf[i])
		}

		inPool.Put(inBuf)
		output <- outBuf
	}
}
