package demodulator

import "sync"

// FmDemodulate will detect FM audio from the input signal. It will keep the same sample rate, so decimation needs to happen before
// This function assumes 48kHz sample rate.
func FmDemodulate(input chan []complex64, inPool *sync.Pool, output chan []int16, outPool *sync.Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	defer close(output)

	var last complex64

	// DC-blocker state
	var xPrev float32
	var yPrev float32
	const dcAlpha = float32(0.995) // ~50–100 Hz @ 48 kHz

	const gain = 5000.0 // audio gain TODO: tune this

	for in := range input {
		out := outPool.Get().([]int16)

		for i := 0; i < len(in); i++ {
			cur := in[i]

			if last != 0 {
				I1, Q1 := real(cur), imag(cur)
				I0, Q0 := real(last), imag(last)

				// Phase-difference FM discriminator
				den := I0*I0 + Q0*Q0
				var v float32
				if den > 1e-12 {
					v = (I1*Q0 - Q1*I0) / den
				} else {
					v = 0
				}

				// --- DC blocker ---
				// y[n]=x[n]−x[n−1]+αy[n−1]
				y := v - xPrev + dcAlpha*yPrev
				xPrev = v
				yPrev = y
				v = y
				// ------------------

				// Gain + clip → int16
				v *= gain
				if v > 1 {
					v = 1
				} else if v < -1 {
					v = -1
				}

				out[i] = int16(v * 32767)
			} else {
				out[i] = 0
			}

			last = cur
		}

		// input-buffer back to pool
		inPool.Put(in)

		// pass output-buffer
		output <- out
	}
}
