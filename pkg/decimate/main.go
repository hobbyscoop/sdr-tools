package main

import (
	"math"
	"sync"
)

// --- Kaiser helpers ---

func kaiserBeta(attDb float64) float64 {
	switch {
	case attDb > 50:
		return 0.1102 * (attDb - 8.7)
	case attDb >= 21:
		return 0.5842*math.Pow(attDb-21, 0.4) + 0.07886*(attDb-21)
	default:
		return 0
	}
}

func besselI0(x float64) float64 {
	sum := 1.0
	y := x * x / 4
	t := y
	for k := 1; k < 25; k++ {
		sum += t
		t *= y / float64(k*k)
	}
	return sum
}

func kaiserWindow(n, N int, beta float64) float64 {
	r := float64(2*n-N) / float64(N)
	return besselI0(beta*math.Sqrt(1-r*r)) / besselI0(beta)
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func designLowpassKaiser(
	numTaps int,
	cutoffHz float64,
	sampleRate float64,
	stopbandDb float64,
) []float32 {

	taps := make([]float32, numTaps)
	beta := kaiserBeta(stopbandDb)
	fc := cutoffHz / sampleRate
	M := numTaps - 1
	mid := float64(M) / 2

	var sum float64

	for n := 0; n < numTaps; n++ {
		x := float64(n) - mid
		h := 2 * fc * sinc(2*fc*x)
		w := kaiserWindow(n, M, beta)
		v := h * w
		taps[n] = float32(v)
		sum += v
	}

	// DC-gain normalisation
	for i := range taps {
		taps[i] /= float32(sum)
	}

	return taps
}

type PolyphaseDecimator struct {
	M      int         // decimate factor
	phases [][]float32 // [M][]
	delay  []complex64
	pos    int
}

// NewPolyphaseDecimator creates a new PolyphaseDecimator.
//
//	  inSampleRate: sample rate of the input signal
//		 decimateFactor: decimation factor (inSampleRate / decimateFactor = output sample rate)
//		 numTaps: number of taps in the FIR filter:
//		   decimateFactor <= 10: 63
//		   decimateFactor 20-50: 127
//	    bad SNR: 191
func NewPolyphaseDecimator(
	inSampleRate float64,
	decimateFactor int,
	numTaps int,
) *PolyphaseDecimator {

	fsOut := inSampleRate / float64(decimateFactor)
	cutoff := 0.45 * fsOut

	taps := designLowpassKaiser(
		numTaps,
		cutoff,
		inSampleRate,
		60, // stopband attenuation
	)

	// Polyphase spliting
	phases := make([][]float32, decimateFactor)
	for i := range phases {
		for j := i; j < len(taps); j += decimateFactor {
			phases[i] = append(phases[i], taps[j])
		}
	}

	maxPhaseLen := 0
	for _, p := range phases {
		if len(p) > maxPhaseLen {
			maxPhaseLen = len(p)
		}
	}

	return &PolyphaseDecimator{
		M:      decimateFactor,
		phases: phases,
		delay:  make([]complex64, maxPhaseLen),
	}
}

// DecimatePolyphase processes input samples through a polyphase decimator, reducing the sampling rate by a factor of M.
// input is a channel receiving chunks of input samples, with each chunk being a slice of complex64 values.
// inPool is a sync.Pool used to manage memory for input slices to reduce garbage collection overhead.
// output is a channel where output decimated sample slices are sent, with each slice being a slice of complex64 values.
// outPool is a sync.Pool used to manage memory for output slices to reduce garbage collection overhead.
// wg is an optional WaitGroup to signal when processing is complete; it is decremented once the method finishes.
func (dec *PolyphaseDecimator) DecimatePolyphase(
	input chan []complex64,
	inPool *sync.Pool,
	output chan []complex64,
	outPool *sync.Pool,
	wg *sync.WaitGroup,
) {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	phase := 0

	for in := range input {
		out := outPool.Get().([]complex64)
		outIdx := 0

		for _, x := range in {
			// shift-register
			dec.delay[dec.pos] = x
			dec.pos++
			if dec.pos == len(dec.delay) {
				dec.pos = 0
			}

			if phase == 0 {
				taps := dec.phases[0]
				var acc complex64
				idx := dec.pos

				for _, t := range taps {
					if idx == 0 {
						idx = len(dec.delay)
					}
					idx--
					acc += complex(
						float32(real(dec.delay[idx]))*t,
						float32(imag(dec.delay[idx]))*t,
					)
				}

				out[outIdx] = acc
				outIdx++
			}

			phase++
			if phase == dec.M {
				phase = 0
			}
		}

		inPool.Put(in)
		output <- out[:outIdx]
	}
}
