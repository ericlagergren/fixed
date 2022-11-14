package fixed

import (
	"errors"
	"math/bits"
	"strconv"
)

const (
	maxUint64Digits = 20
	maxUint96Digits = 29
)

var pow10tab = [...]uint64{
	1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
}

// digits returns the number of decimal digits in x.
func digits(v uint64) int {
	if v < 10 {
		return 1
	}
	// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
	t := (bits.Len64(v) * 1233) / 4096
	if v < pow10tab[t] {
		return t
	}
	return t + 1
}

func lower(c byte) byte {
	return c | ('x' - 'X')
}

func syntaxError(fn, str string) *strconv.NumError {
	return &strconv.NumError{
		Func: fn,
		Num:  cloneString(str),
		Err:  strconv.ErrSyntax,
	}
}

func rangeError(fn, str string) *strconv.NumError {
	return &strconv.NumError{
		Func: fn,
		Num:  cloneString(str),
		Err:  strconv.ErrRange,
	}
}

func baseError(fn, str string, base int) *strconv.NumError {
	return &strconv.NumError{
		Func: fn,
		Num:  cloneString(str),
		Err:  errors.New("invalid base " + strconv.Itoa(base)),
	}
}

func cloneString(x string) string {
	return string([]byte(x))
}
