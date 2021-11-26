package main

// import (
// 	"fmt"
// 	"math"
// )

// func main() {
// 	xs := []complex64{
// 		complex(float32(0.1), float32(0.9)),
// 		complex(float32(0.2), float32(0.8)),
// 		complex(float32(0.3), float32(0.7)),
// 		complex(float32(0.4), float32(0.6)),
// 		complex(float32(0.5), float32(0.5)),
// 	}

// 	for _, x := range complex64sToIQSamples(xs) {
// 		fmt.Println(int8(x))
// 	}

// 	for _, x := range iqSamplesToComplex64s(complex64sToIQSamples(xs)) {
// 		fmt.Println(real(x), imag(x))
// 	}
// }

// func complex64sToIQSamples(c64s []complex64) []byte {
// 	i8s := make([]byte, len(c64s)*2)
// 	for i := 0; i < len(c64s); i++ {
// 		i8s[i*2] = byte(int8(real(c64s[i]) * math.MaxInt8))
// 		i8s[i*2+1] = byte(int8(imag(c64s[i]) * math.MaxInt8))
// 	}
// 	return i8s
// }

// func iqSamplesToComplex64s(i8s []byte) []complex64 {
// 	c64s := make([]complex64, len(i8s)/2)
// 	for i := 0; i < len(i8s)/2; i++ {
// 		c64s[i] = complex(float32(int8(i8s[i*2]))/math.MaxInt8, float32(int8(i8s[i*2+1]))/math.MaxInt8)
// 	}
// 	return c64s
// }
