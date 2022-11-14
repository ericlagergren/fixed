package fixed

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

var maxUint128 = Uint128{
	math.MaxUint64,
	math.MaxUint64,
}

// Uint128 is an unsigned, 128-bit integer.
//
// It can be compared for equality with ==.
type Uint128 struct {
	u0, u1 uint64
}

var _ Uint[Uint128] = Uint128{}

// U128 returns x as a Uint128.
func U128(x uint64) Uint128 {
	return Uint128{x, 0}
}

func (Uint128) max() Uint128 {
	return Uint128{
		math.MaxUint64,
		math.MaxUint64,
	}
}

func (x Uint128) uint96() Uint96 {
	return Uint96{x.u0, uint32(x.u1)}
}

func (x Uint128) uint192() Uint192 {
	return Uint192{x.u0, x.u1, 0}
}

func (x Uint128) uint256() Uint256 {
	return Uint256{x.u0, x.u1, 0, 0}
}

// BitLen returns the number of bits required to represent x.
func (x Uint128) BitLen() int {
	if x.u1 != 0 {
		return 64 + bits.Len64(x.u1)
	}
	return bits.Len64(x.u0)
}

// LeadingZeros returns the number of leading zeros in x.
func (x Uint128) LeadingZeros() int {
	return 128 - x.BitLen()
}

// IsZero is shorthand for x == Uint128{}.
func (x Uint128) IsZero() bool {
	return x == Uint128{}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint128) Cmp(y Uint128) int {
	switch {
	case x == y:
		return 0
	case x.u1 < y.u1, x.u1 == y.u1 && x.u0 < y.u0:
		return -1
	default:
		return +1
	}
}

// cmp64 compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint128) cmp64(y uint64) int {
	if x.u1 != 0 {
		return +1
	}
	switch x := x.u0; {
	case x > y:
		return +1
	case x < y:
		return -1
	default:
		return 0
	}
}

// And returns x&y.
func (x Uint128) And(y Uint128) Uint128 {
	return Uint128{x.u0 & y.u0, x.u1 & y.u1}
}

// Or returns x|y.
func (x Uint128) Or(y Uint128) Uint128 {
	return Uint128{x.u0 | y.u0, x.u1 | y.u1}
}

// Xor returns x^y.
func (x Uint128) Xor(y Uint128) Uint128 {
	return Uint128{x.u0 ^ y.u0, x.u1 ^ y.u1}
}

// Lsh returns x<<n.
func (x Uint128) Lsh(n uint) Uint128 {
	if n > 64 {
		return Uint128{0, x.u0 << (n - 64)}
	}
	return Uint128{x.u0 << n, x.u1<<n | x.u0>>(64-n)}
}

// Rsh returns x>>n.
func (x Uint128) Rsh(n uint) Uint128 {
	if n > 64 {
		return Uint128{x.u1 >> (n - 64), 0}
	}
	return Uint128{x.u0>>n | x.u1<<(64-n), x.u1 >> n}
}

// Add returns x+y.
func (x Uint128) Add(y Uint128) Uint128 {
	z, _ := x.AddCheck(y)
	return z
}

// add64 returns x+y.
func (x Uint128) add64(y uint64) Uint128 {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, _ := bits.Add64(x.u1, 0, c)
	return Uint128{u0, u1}
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint128) AddCheck(y Uint128) (z Uint128, carry uint64) {
	u0, c := bits.Add64(x.u0, y.u0, 0)
	u1, c := bits.Add64(x.u1, y.u1, c)
	return Uint128{u0, u1}, c
}

// add64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint128) addCheck64(y uint64) (z Uint128, carry uint64) {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, c := bits.Add64(x.u1, 0, c)
	return Uint128{u0, u1}, c
}

// Sub returns x-y.
func (x Uint128) Sub(y Uint128) Uint128 {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, _ := bits.Sub64(x.u1, y.u1, b)
	return Uint128{u0, u1}
}

func (x Uint128) sub64(y uint64) Uint128 {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, _ := bits.Sub64(x.u1, 0, b)
	return Uint128{u0, u1}
}

// SubCheck returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x Uint128) SubCheck(y Uint128) (z Uint128, borrow uint64) {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, b := bits.Sub64(x.u1, y.u1, b)
	return Uint128{u0, u1}, b
}

func (x Uint128) subCheck64(y uint64) (z Uint128, borrow uint64) {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, b := bits.Sub64(x.u1, 0, b)
	return Uint128{u0, u1}, b
}

// Mul returns x*y.
func (x Uint128) Mul(y Uint128) Uint128 {
	u1, u0 := bits.Mul64(x.u0, y.u0)
	return Uint128{u0, u1 + x.u1*y.u0 + x.u0*y.u1}
}

// MulCheck returns x*y and reports whether the multiplication
// oveflowed.
func (x Uint128) MulCheck(y Uint128) (Uint128, bool) {
	if x.BitLen()+y.BitLen() > 128 {
		return Uint128{}, false
	}

	var u0, u1 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		if c != 0 {
			return Uint128{}, false
		}
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		if c != 0 {
			return Uint128{}, false
		}
	}
	return Uint128{u0, u1}, true
}

func (x Uint128) mul64(y uint64) Uint128 {
	hi, lo := bits.Mul64(x.u0, y)
	return Uint128{lo, hi + x.u1*y}
}

func (x Uint128) mulCheck64(y uint64) (Uint128, bool) {
	if y == 0 {
		return Uint128{}, true
	}
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(x.u1, y, c)
	if c != 0 {
		return Uint128{}, false
	}
	return Uint128{u0, u1}, true
}

// QuoRem returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint128) QuoRem(y Uint128) (q, r Uint128) {
	if x.Cmp(y) < 0 {
		// x/y for x < y = 0.
		// x%y for x < y = x.
		return Uint128{}, x
	}
	if y.u1 == 0 {
		// Fast path for a 64-bit y.
		q, r64 := x.quoRem64(y.u0)
		return q, U128(r64)
	}

	n := uint(bits.LeadingZeros64(y.u1))
	y1 := y.Lsh(n)
	x1 := x.Rsh(1)
	tq, _ := bits.Div64(x1.u1, x1.u0, y1.u1)
	tq >>= 63 - n
	if tq != 0 {
		tq--
	}
	q = U128(tq)
	ytq := y.mul64(tq) // ytq := y*tq
	r = x.Sub(ytq)     // r = x-ytq
	if r.Cmp(y) >= 0 {
		q = q.add64(1) // q++
		r = r.Sub(y)   // r -= y
	}
	return
}

// quoRem64 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint128) quoRem64(y uint64) (q Uint128, r uint64) {
	if x.u1 < y {
		lo, r := bits.Div64(x.u1, x.u0, y)
		return Uint128{lo, 0}, r
	}
	hi, r := bits.Div64(0, x.u1, y)
	lo, r := bits.Div64(r, x.u0, y)
	return Uint128{lo, hi}, r
}

// div128 returns (q, r) such that
//
//  q = (hi, lo)/y
//  r = (hi, lo) - y*q
func div128(hi, lo, y Uint128) (q, r Uint128) {
	if y.IsZero() {
		panic("integer divide by zero")
	}
	if y.Cmp(hi) <= 0 {
		panic("integer overflow")
	}

	s := uint(y.LeadingZeros())
	y = y.Lsh(s) // y = y<<s
	yn1 := y.u1  // yn1 := y >> 64
	yn0 := y.u0  // yn0 := y & mask64

	un32 := hi.Lsh(s).Or(lo.Rsh(128 - s)) // un32 := hi<<s | lo>>(128-s)
	un10 := lo.Lsh(s)                     // un10 := lo<<s
	un1 := un10.u1                        // un1 := un10 >> 64
	un0 := un10.u0                        // un0 := un10 & mask64
	q1, rhat := un32.quoRem64(yn1)

	var c uint64 // rhat + yn1 carry

	// for q1 >= two64 || q1*yn0 > two64*rhat+un1 { ... }
	for q1.u1 != 0 || q1.mul64(yn0).Cmp(Uint128{un1, rhat}) > 0 {
		q1 = q1.sub64(1)                   // q1--
		rhat, c = bits.Add64(rhat, yn1, 0) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// un21 := un32*two64 + un1 - q1*y
	un21 := Uint128{un1, un32.u0}.Sub(q1.Mul(y))
	q0, rhat := un21.quoRem64(yn1)

	// for q0 >= two64 || q0*yn0 > two64*rhat+un0 { ... }
	for q0.u1 != 0 || q0.mul64(yn0).Cmp(Uint128{un0, rhat}) > 0 {
		q0 = q0.sub64(1)                   // q0--
		rhat, c = bits.Add64(rhat, yn1, 0) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// q = q1*two64 + q0
	q = Uint128{q0.u0, q1.u0}
	// r = (un21*two64 + un0 - q0*y) >> s
	r = Uint128{un0, un21.u0}.Sub(q0.Mul(y)).Rsh(s)
	return
}

func (x Uint128) GoString() string {
	return fmt.Sprintf("[%d %d]", x.u0, x.u1)
}

// String returns the base-10 representation of x.
func (x Uint128) String() string {
	if x.u1 == 0 {
		return strconv.FormatUint(x.u0, 10)
	}
	b := make([]byte, 39)
	i := len(b)
	for x.cmp64(10) >= 0 {
		q, r := x.quoRem64(10)
		i--
		b[i] = byte(r + '0')
		x = q
	}
	i--
	b[i] = byte(x.u0 + '0')
	return string(b[i:])
}

// ParseUint128 returns the value of s in the given base.
func ParseUint128(s string, base int) (Uint128, error) {
	x, _, _, err := parseUint[Uint128](s, base, false)
	return x, err
}
