# sdr-tools
Various tools to use with SDR I/Q streams

## Tools
These tools are written to consume from stdin and produce to stdout.
```bash
$ rtl-tcp -f 144.8M| convert -i char -o float | firdecimate -d 20 -r 0.000001 | fmdemodulator -n > output.pcm
```

### Convert
Converts between different data types.
Supported formats:

| *in*      | *out*       | *usage*                                          |
|-----------|-------------|--------------------------------------------------|
| `int8`    | `complex64` | reading from rtl-sdr                             |
| `float32` | `int16`     | writing to audio (sox -t raw -c1 -b16 -e signed) |

### shift

### fir decimate

### fm demodulator
Demodulates FM signals from complex I/Q samples to audio samples.
It uses the same sample rate as the input samples. And applies a simple DC blocking filter.

### gfsk demodulator

