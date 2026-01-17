package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/hobbyscoop/sdr-tools/pkg/convert"
	"github.com/hobbyscoop/sdr-tools/pkg/io"
)

var (
	inType     = flag.String("in", "int8", "input sample data type")
	outType    = flag.String("out", "complex64", "output sample data type")
	bufferSize = flag.Int("buffer", 4096, "buffer size for reading/writing in samples")
)

func main() {
	flag.Parse()
	wg := &sync.WaitGroup{}
	convertFound := false

	if *inType == "int8" && *outType == "complex64" {
		convertFound = true

		inPool := sync.Pool{
			New: func() any {
				return make([]byte, *bufferSize*2)
			},
		}
		input := make(chan []byte, 4)

		outPool := sync.Pool{
			New: func() any {
				return make([]complex64, *bufferSize)
			},
		}
		output := make(chan []complex64, 4)

		wg.Add(3)
		go convert.Int8ToComplex64(input, &inPool, output, &outPool, wg)

		go func() {
			err := io.ByteReader(input, &inPool, wg)
			if err != nil {
				panic(err)
			}
		}()

		go func() {
			err := io.Complex64Writer(output, &outPool, wg)
			if err != nil {
				panic(err)
			}
		}()
	}

	if *inType == "float32" && *outType == "int16" {
		convertFound = true
		inPool := sync.Pool{
			New: func() any {
				return make([]float32, *bufferSize*2)
			},
		}
		input := make(chan []float32, 4)

		outPool := sync.Pool{
			New: func() any {
				return make([]int16, *bufferSize)
			},
		}
		output := make(chan []int16, 4)

		wg.Add(3)
		go convert.Float32ToInt16(input, &inPool, output, &outPool, wg)

		go func() {
			err := io.Float32Reader(input, &inPool, wg)
			if err != nil {
				panic(err)
			}
		}()

		go func() {
			err := io.Int16Writer(output, &outPool, wg)
			if err != nil {
				panic(err)
			}
		}()
	}
	// TODO: handle other types

	if !convertFound {
		fmt.Println("unsupported conversion")
		// TODO: explain supported conversions
		os.Exit(1)
	}

	wg.Wait()
}

func usage() {
	fmt.Println("usage: convert -in <input type> -out <output type>")
	os.Exit(1)
}
