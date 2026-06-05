package mandelbrot

import (
	"fmt"
	"strconv"
)

func formatComplex(c complex128) string {
	realpart := strconv.FormatFloat(real(c), 'f', -1, 64)
	imagPart := strconv.FormatFloat(imag(c), 'f', -1, 64)
	return fmt.Sprintf("(%s+%si)", realpart, imagPart)
}

func funcMandel(z complex64, c complex64) complex64 {
	return z*z + c
}

func mandelIter(z complex64, c complex64, maxIter int) int {
	for i := 0; i < maxIter; i++ {
		if real(z)*real(z)+imag(z)*imag(z) > 4 {
			return i
		}
		z = funcMandel(z, c)
	}
	return maxIter
}
