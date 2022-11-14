package fixed

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Dec128 is a signed, 128-bit decimal.
//
// It has the form coeff * 10^exp where coeff supports the range
// [-(2⁹⁶-1), 2⁹⁶-1] and exp supports the range [-28, 28].
//
// Dec128 values should only be compared with Cmp or Equal.
// Comparing with == could provide incorrect results. For
// example, "12.00" == "12" will compare false.
type Dec128 struct {
	coeff Uint96
	flags flags
}

const (
	signbit  = 0x80000000
	expmask  = 0x0000ffff
	expLimit = 10 * maxUint96Digits // E_limit
	maxExp   = expLimit             // E_max
	minExp   = -maxExp + 1          // E_min
)

type flags uint32

func makeFlags(neg bool, exp int) flags {
	if exp > maxExp || exp < minExp {
		panic("exp out of range: " + strconv.Itoa(exp))
	}
	f := flags(uint16(exp)) & expmask
	if neg {
		f |= signbit
	}
	return f
}

func (f flags) exp() int {
	return int(int16(f & expmask))
}

func (f flags) sign() int {
	return int((f & signbit) >> 31)
}

func (f flags) signbit() bool {
	return f.sign() == 1
}

func (f *flags) setExp(exp int) {
	*f |= flags(uint16(exp)) & expmask
}

// exp returns the decimal's exponent.
func (x Dec128) exp() int {
	return x.flags.exp()
}

func (x Dec128) signbit() bool {
	return x.flags.signbit()
}

func (z *Dec128) setExp(exp int) {
	z.flags.setExp(exp)
}

// adjExp returns the decimal's adjusted exponent.
func (x Dec128) adjExp() int {
	return x.exp() + (x.coeff.digits() - 1)
}

func (x Dec128) max() Dec128 {
	return Dec128{
		Uint96{}.max(),
		makeFlags(false, maxExp),
	}
}

func (x Dec128) min() Dec128 {
	return Dec128{
		Uint96{}.max(),
		makeFlags(true, maxExp),
	}
}

// Abs returns the absolute value of x.
func (x Dec128) Abs() Dec128 {
	return Dec128{x.coeff, x.flags &^ signbit}
}

// Add returns x+y.
func (x Dec128) Add(y Dec128) Dec128 {
	fmt.Printf("Add: %s + %s\n", x, y)
	flags := x.flags ^ y.flags
	if flags&expmask != 0 {
		d := x.exp() - y.exp()

		// Align the coeffecients.
		//
		// Try and do it in 96 bits. If we can't, fall back to
		// the slow path.
		var coeff Uint96
		var ok bool
		if d >= 0 {
			coeff, ok = x.coeff.mulPow10(uint(d))
		} else {
			coeff, ok = y.coeff.mulPow10(uint(-d))
		}
		if !ok {
			return x.addSlow(y)
		}

		if d >= 0 {
			x = Dec128{coeff, makeFlags(x.signbit(), y.exp())}
		} else {
			y = Dec128{coeff, makeFlags(y.signbit(), x.exp())}
		}
	}

	if flags&signbit == 0 {
		// x + y = x + y
		// (-x) + (-y) = -(x + y)
		z, c := x.coeff.AddCheck(y.coeff)
		if c != 0 {
			panic("overflow")
		}
		return Dec128{z, x.flags}
	}

	// x + (-y) = x - y = -(y - x)
	// (-x) + y = y - x = -(x - y)
	z, b := x.coeff.SubCheck(y.coeff)
	if b != 0 {
		// x < y, so the result is negative.
		x.flags ^= signbit
		z = y.coeff.Sub(x.coeff)
	}
	return Dec128{z, x.flags}
}

func (x Dec128) addSlow(y Dec128) Dec128 {
	// fmt.Printf("addSlow: %s + %s (%d - %d, %d)\n",
	// 	x, y, x.exp(), y.exp(), x.exp()-y.exp())

	lhs := x.coeff.uint256()
	rhs := y.coeff.uint256()

	var ok bool
	if d := x.exp() - y.exp(); d >= 0 {
		//fmt.Printf("lhs = %s * 10^%d\n", lhs, uint(d))
		lhs, ok = lhs.mulPow10(uint(d))
	} else {
		//fmt.Printf("rhs = %s * 10^%d\n", rhs, uint(-d))
		rhs, ok = rhs.mulPow10(uint(-d))
	}
	if !ok {
		// Can't even fit into 256 bits, so fall back to the
		// slowest path.
		return x.addSlower(y)
	}

	flags := x.flags
	flags.setExp(min(x.exp(), y.exp()))

	// x + (-y) = x - y = -(y - x)
	// (-x) + y = y - x = -(x - y)
	z, b := lhs.SubCheck(rhs)
	if b != 0 {
		// x < y, so the result is negative.
		flags ^= signbit
		z = rhs.Sub(lhs)
	}
	if z.u2 != 0 || z.u1 > math.MaxUint32 {
		//panic("overflow")
	}
	return Dec128{z.low96(), flags}
}

func (x Dec128) addSlower(y Dec128) Dec128 {
	// fmt.Printf("addSlower: %s + %s (%d - %d, %d)\n",
	// 	x, y, x.exp(), y.exp(), x.exp()-y.exp())

	var lhs, rhs big.Int
	lhs.SetBits(x.coeff.words())
	rhs.SetBits(y.coeff.words())

	if d := x.exp() - y.exp(); d >= 0 {
		//fmt.Printf("lhs = %s * 10^%d\n", &lhs, uint(d))
		bigMulPow10(&lhs, uint(d))
	} else {
		//fmt.Printf("rhs = %s * 10^%d\n", &rhs, uint(-d))
		bigMulPow10(&rhs, uint(-d))
	}

	flags := x.flags
	flags.setExp(min(x.exp(), y.exp()))

	var res big.Int
	//res.SetBits(make([]big.Word, 0, 14))
	// x + (-y) = x - y = -(y - x)
	// (-x) + y = y - x = -(x - y)
	if lhs.Sub(&res, &rhs).Sign() < 0 {
		// x < y, so the result is negative.
		flags ^= signbit
		rhs.Sub(&res, &lhs)
	}
	z, ok := u96FromBig(&res)
	if !ok {
		panic("overflow")
	}
	return Dec128{z, flags}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Dec128) Cmp(y Dec128) int {
	flags := x.flags ^ y.flags

	if flags&signbit != 0 {
		// Signs differ.
		if x.flags&signbit != 0 {
			return -1
		}
		return +1
	}

	if flags&expmask == 0 {
		// Same exponent.
		return x.coeff.Cmp(y.coeff)
	}

	// TODO(eric): exp
	return 0
}

// Equal reports whether x is equal to y.
func (x Dec128) Equal(y Dec128) bool {
	return x.Cmp(y) == 0
}

func (x Dec128) GoString() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%d, %s, %d]",
		x.flags.sign(), x.coeff, x.exp())
	return b.String()
}

// Neg returns -x.
func (x Dec128) Neg() Dec128 {
	return Dec128{x.coeff, x.flags ^ signbit}
}

// Sign returns
//
//   - +1 if x > 0
//   - 0 if x == 0
//   - -1 if x < 0
func (x Dec128) Sign() int {
	switch {
	case x.signbit():
		return -1
	case x.coeff != Uint96{}:
		return +1
	default:
		return 0
	}
}

// String returns the decimal formatted as a scientific string.
func (x Dec128) String() string {
	exp := x.exp()
	adj := exp + (x.coeff.digits() - 1)
	if exp <= 0 && adj >= -6 {
		// Without exponential notation: -ddddd.dddd
		if exp == 0 {
			// Easy case: no decimal point.
			b := make([]byte, 1, 1+maxUint96Digits)
			b = x.coeff.append(b)
			if x.flags&signbit != 0 {
				b[0] = '-'
			} else {
				b = b[1:]
			}
			return string(b)
		}

		s := make([]byte, 0, maxUint96Digits)
		s = x.coeff.append(s)

		// pad is the maximum number of zeros that we might need
		// for padding.
		//
		// s is at most maxUint96Digits and exp is non-zero.
		const padDigits = maxUint96Digits - 1
		b := make([]byte, 0, 1+maxUint96Digits+1+padDigits)
		if x.flags&signbit != 0 {
			b = append(b, '-')
		}

		// We want
		//   s = s[:len(s)-abs] + "." + s[len(s)-abs:]
		//   if s[0] == '.' { s += "0" }
		abs := -exp
		if len(s) > abs {
			b = append(b, s[:len(s)-abs]...)
		} else {
			b = append(b, '0')
		}
		b = append(b, '.')
		for abs > len(s) {
			b = append(b, '0')
			abs--
		}
		b = append(b, s[len(s)-abs:]...)
		return string(b)
	}

	// With exponential notation: -d.ddde±ddd
	b := make([]byte, 2, 2+maxUint96Digits+2+maxUint64Digits)

	i := 2
	b = x.coeff.append(b)
	if len(b) > 2+1 {
		// We have
		//   [ 0 0 d d d ... ]
		// but we want
		//   [ 0 d . d d ... ]
		b[1], b[2] = b[2], '.'
		i--
	}
	b = append(b, 'e')
	if adj > 0 {
		b = append(b, '+')
	}
	b = strconv.AppendInt(b, int64(adj), 10)
	if x.flags&signbit != 0 {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// Sub returns x-y.
func (x Dec128) Sub(y Dec128) Dec128 {
	return x.Add(y.Neg())
}

// ParseDec128 returns the value of s in the given base.
func ParseDec128(s string, base int) (Dec128, error) {
	const fnParseDec = "ParseDec128"

	if s == "" {
		return Dec128{}, syntaxError(fnParseDec, s)
	}

	neg := s[0] == '-'
	if neg {
		s = s[1:]
	}

	// TODO(eric): inf, NaN, etc.

	coeff, expIdx, dotIdx, err := parseUint[Uint96](s, base, true)
	if err != nil {
		if e, ok := err.(*strconv.NumError); ok {
			e.Func = fnParseDec
		}
		return Dec128{}, err
	}

	var exp int
	if expIdx < len(s) {
		exp, err = parseExp(s[expIdx:], fnParseDec)
		if err != nil {
			return Dec128{}, err
		}
	}
	if dotIdx > 0 {
		exp -= expIdx - 1 - dotIdx
	}
	x := Dec128{coeff, makeFlags(neg, exp)}
	if exp := x.adjExp(); exp < minExp || exp > maxExp {
		println(exp, minExp, maxExp)
		return Dec128{}, rangeError(fnParseDec, s)
	}
	return x, nil
}

func parseExp(s, fn string) (int, error) {
	if s == "" {
		return 0, syntaxError(fn, s)
	}
	switch s[0] {
	case 'e', 'E':
		return strconv.Atoi(s[1:])
	default:
		return 0, syntaxError(fn, s)
	}
}
