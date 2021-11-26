package main

import (
	"math"
)

const (
	Hz  = 1
	KHz = 1000 * Hz
	MHz = 1000 * KHz
	GHz = 1000 * MHz
)

func complex64sToIQSamples(c64s []complex64) []byte {
	i8s := make([]byte, len(c64s)*2)
	for i := 0; i < len(c64s); i++ {
		i8s[i*2] = byte(int8(real(c64s[i]) * math.MaxInt8))
		i8s[i*2+1] = byte(int8(imag(c64s[i]) * math.MaxInt8))
	}
	return i8s
}

func iqSamplesToComplex64s(i8s []byte) []complex64 {
	c64s := make([]complex64, len(i8s)/2)
	for i := 0; i < len(i8s)/2; i++ {
		c64s[i] = complex(float32(int8(i8s[i*2]))/math.MaxInt8, float32(int8(i8s[i*2+1]))/math.MaxInt8)
	}
	return c64s
}
