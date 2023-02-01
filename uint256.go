// Code generated by 'gen'. DO NOT EDIT.

package fixed

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"sync"
)

// Uint256 is an unsigned, 256-bit integer.
//
// It can be compared for equality with ==.
type Uint256 struct {
	u0, u1, u2, u3 uint64
}

var _ Uint[Uint256] = Uint256{}

// U256 returns x as a Uint256.
func U256(x uint64) Uint256 {
	return Uint256{u0: x}
}

func u256(lo, hi Uint128) Uint256 {
	return Uint256{lo.u0, lo.u1, hi.u0, hi.u1}
}

func (Uint256) max() Uint256 {
	return Uint256{
		math.MaxUint64,
		math.MaxUint64,
		math.MaxUint64,
		math.MaxUint64,
	}
}

func (x Uint256) low() Uint128 {
	return Uint128{x.u0, x.u1}
}

func (x Uint256) high() Uint128 {
	return Uint128{x.u2, x.u3}
}

func (x Uint256) uint8() uint8 {
	return uint8(x.u0)
}

// Bytes encodes x as a little-endian integer.
func (x Uint256) Bytes(b *[32]byte) {
	binary.LittleEndian.PutUint64(b[0:], x.u0)
	binary.LittleEndian.PutUint64(b[8:], x.u1)
	binary.LittleEndian.PutUint64(b[16:], x.u2)
	binary.LittleEndian.PutUint64(b[24:], x.u3)

}

// SetBytes sets x to the encoded little-endian integer b.
func (x *Uint256) SetBytes(b []byte) error {
	if len(b) != 32 {
		return fmt.Errorf("fixed: invalid length: %d", len(b))
	}
	x.u0 = binary.LittleEndian.Uint64(b[0:])
	x.u1 = binary.LittleEndian.Uint64(b[8:])
	x.u2 = binary.LittleEndian.Uint64(b[16:])
	x.u3 = binary.LittleEndian.Uint64(b[24:])
	return nil
}

// Size returns the width of the integer in bits.
func (Uint256) Size() int {
	return 256
}

// BitLen returns the number of bits required to represent x.
func (x Uint256) BitLen() int {
	switch {
	case x.u3 != 0:
		return 192 + bits.Len64(x.u3)
	case x.u2 != 0:
		return 128 + bits.Len64(x.u2)
	case x.u1 != 0:
		return 64 + bits.Len64(x.u1)
	default:
		return bits.Len64(x.u0)
	}
}

// LeadingZeros returns the number of leading zeros in x.
func (x Uint256) LeadingZeros() int {
	return 256 - x.BitLen()
}

// IsZero is shorthand for x == Uint256{}.
func (x Uint256) IsZero() bool {
	return x == Uint256{}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint256) Cmp(y Uint256) int {
	switch {
	case x.u3 != y.u3:
		return cmp(x.u3, y.u3)
	case x.u2 != y.u2:
		return cmp(x.u2, y.u2)
	case x.u1 != y.u1:
		return cmp(x.u1, y.u1)
	default:
		return cmp(x.u0, y.u0)
	}
}

// cmp64 compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint256) cmp64(y uint64) int {
	v := x
	v.u0 = 0
	if !v.IsZero() {
		return +1
	}
	return cmp(x.u0, y)
}

// Equal reports whether x == y.
//
// In general, prefer the == operator to using this method.
func (x Uint256) Equal(y Uint256) bool {
	return x == y
}

// And returns x&y.
func (x Uint256) And(y Uint256) Uint256 {
	return Uint256{
		x.u0 & y.u0,
		x.u1 & y.u1,
		x.u2 & y.u2,
		x.u3 & y.u3,
	}
}

// Or returns x|y.
func (x Uint256) Or(y Uint256) Uint256 {
	return Uint256{
		x.u0 | y.u0,
		x.u1 | y.u1,
		x.u2 | y.u2,
		x.u3 | y.u3,
	}
}

// orLsh64 returns x | y<<s.
func (x Uint256) orLsh64(y uint64, s uint) Uint256 {
	return x.Or(Uint256{u0: y}.Lsh(s))
}

// Xor returns x^y.
func (x Uint256) Xor(y Uint256) Uint256 {
	return Uint256{
		x.u0 ^ y.u0,
		x.u1 ^ y.u1,
		x.u2 ^ y.u2,
		x.u3 ^ y.u3,
	}
}

// Lsh returns x<<n.
func (x Uint256) Lsh(n uint) Uint256 {
	switch {
	case n > 192:
		s := n - 192
		return Uint256{
			u3: x.u0 << s,
		}
	case n > 128:
		s := n - 128
		ŝ := 64 - s
		return Uint256{
			u2: x.u0 << s,
			u3: x.u1<<s | x.u0>>ŝ,
		}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint256{
			u1: x.u0 << s,
			u2: x.u1<<s | x.u0>>ŝ,
			u3: x.u2<<s | x.u1>>ŝ,
		}
	default:
		s := n
		ŝ := 64 - s
		return Uint256{
			u0: x.u0 << s,
			u1: x.u1<<s | x.u0>>ŝ,
			u2: x.u2<<s | x.u1>>ŝ,
			u3: x.u3<<s | x.u2>>ŝ,
		}
	}
}

// Rsh returns x>>n.
func (x Uint256) Rsh(n uint) Uint256 {
	switch {
	case n > 192:
		s := n - 192
		return Uint256{
			u0: x.u3 >> s,
		}
	case n > 128:
		s := n - 128
		ŝ := 64 - s
		return Uint256{
			u0: x.u2>>s | x.u3<<ŝ,
			u1: x.u3 >> s,
		}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint256{
			u0: x.u1>>s | x.u2<<ŝ,
			u1: x.u2>>s | x.u3<<ŝ,
			u2: x.u3 >> s,
		}
	default:
		s := n
		ŝ := 64 - s
		return Uint256{
			u0: x.u0>>s | x.u1<<ŝ,
			u1: x.u1>>s | x.u2<<ŝ,
			u2: x.u2>>s | x.u3<<ŝ,
			u3: x.u3 >> s,
		}
	}
}

// Add returns x+y.
func (x Uint256) Add(y Uint256) Uint256 {
	var z Uint256
	var carry uint64
	z.u0, carry = bits.Add64(x.u0, y.u0, 0)
	z.u1, carry = bits.Add64(x.u1, y.u1, carry)
	z.u2, carry = bits.Add64(x.u2, y.u2, carry)
	z.u3, _ = bits.Add64(x.u3, y.u3, carry)
	return z
}

// add64 returns x+y.
func (x Uint256) add64(y uint64) Uint256 {
	var z Uint256
	var carry uint64
	z.u0, carry = bits.Add64(x.u0, y, 0)
	z.u1, carry = bits.Add64(x.u1, 0, carry)
	z.u2, carry = bits.Add64(x.u2, 0, carry)
	z.u3, _ = bits.Add64(x.u3, 0, carry)
	return z
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint256) AddCheck(y Uint256) (z Uint256, carry uint64) {
	z.u0, carry = bits.Add64(x.u0, y.u0, 0)
	z.u1, carry = bits.Add64(x.u1, y.u1, carry)
	z.u2, carry = bits.Add64(x.u2, y.u2, carry)
	z.u3, carry = bits.Add64(x.u3, y.u3, carry)
	return z, carry
}

// addCheck64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint256) addCheck64(y uint64) (z Uint256, carry uint64) {
	z.u0, carry = bits.Add64(x.u0, y, 0)
	z.u1, carry = bits.Add64(x.u1, 0, carry)
	z.u2, carry = bits.Add64(x.u2, 0, carry)
	z.u3, carry = bits.Add64(x.u3, 0, carry)
	return z, carry
}

// Sub returns x-y.
func (x Uint256) Sub(y Uint256) Uint256 {
	var z Uint256
	var borrow uint64
	z.u0, borrow = bits.Sub64(x.u0, y.u0, 0)
	z.u1, borrow = bits.Sub64(x.u1, y.u1, borrow)
	z.u2, borrow = bits.Sub64(x.u2, y.u2, borrow)
	z.u3, _ = bits.Sub64(x.u3, y.u3, borrow)
	return z
}

// sub64 returns x-y.
func (x Uint256) sub64(y uint64) Uint256 {
	var z Uint256
	var borrow uint64
	z.u0, borrow = bits.Sub64(x.u0, y, 0)
	z.u1, borrow = bits.Sub64(x.u1, 0, borrow)
	z.u2, borrow = bits.Sub64(x.u2, 0, borrow)
	z.u3, _ = bits.Sub64(x.u3, 0, borrow)
	return z
}

// SubCheck returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x Uint256) SubCheck(y Uint256) (z Uint256, borrow uint64) {
	z.u0, borrow = bits.Sub64(x.u0, y.u0, 0)
	z.u1, borrow = bits.Sub64(x.u1, y.u1, borrow)
	z.u2, borrow = bits.Sub64(x.u2, y.u2, borrow)
	z.u3, borrow = bits.Sub64(x.u3, y.u3, borrow)
	return z, borrow
}

// subCheck64 returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x Uint256) subCheck64(y uint64) (z Uint256, borrow uint64) {
	z.u0, borrow = bits.Sub64(x.u0, y, 0)
	z.u1, borrow = bits.Sub64(x.u1, 0, borrow)
	z.u2, borrow = bits.Sub64(x.u2, 0, borrow)
	z.u3, borrow = bits.Sub64(x.u3, 0, borrow)
	return z, borrow
}

// Mul returns x*y.
func (x Uint256) Mul(y Uint256) Uint256 {
	var z Uint256
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, z.u0 = bits.Mul64(x.u0, d)
		c, z.u1 = mulAddWWW(x.u1, d, c)
		c, z.u2 = mulAddWWW(x.u2, d, c)
		z.u3 += x.u3*d + c
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, z.u1 = mulAddWWW(x.u0, d, z.u1)
		c, z.u2 = mulAddWWWW(x.u1, d, z.u2, c)
		z.u3 += x.u2*d + c
	}

	// y.u2 * x
	if d := y.u2; d != 0 {
		c, z.u2 = mulAddWWW(x.u0, d, z.u2)
		z.u3 += x.u1*d + c
	}

	// y.u3 * x
	if d := y.u3; d != 0 {
		z.u3 += x.u0 * d
	}

	return z
}

func (x Uint256) mul128(y Uint128) Uint256 {
	var z Uint256
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, z.u0 = bits.Mul64(x.u0, d)
		c, z.u1 = mulAddWWW(x.u1, d, c)
		c, z.u2 = mulAddWWW(x.u2, d, c)
		z.u3 += x.u3*d + c
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, z.u1 = mulAddWWW(x.u0, d, z.u1)
		c, z.u2 = mulAddWWWW(x.u1, d, z.u2, c)
		z.u3 += x.u2*d + c
	}

	return z
}

func (x Uint256) mul64(y uint64) Uint256 {
	if y == 0 {
		return Uint256{}
	}
	var z Uint256
	var c uint64
	c, z.u0 = bits.Mul64(x.u0, y)
	c, z.u1 = mulAddWWW(x.u1, y, c)
	c, z.u2 = mulAddWWW(x.u2, y, c)
	z.u3 += x.u3*y + c
	return z
}

// Exp return x^y mod m.
//
// If m == 0, Exp simply returns x^y.
func (x Uint256) Exp(y, m Uint256) Uint256 {
	const mask = 1 << (64 - 1)

	// x^0 = 1
	if y.IsZero() {
		return U256(1)
	}

	// x^1 mod m == x mod m
	mod := !m.IsZero()
	if y == U256(1) && mod {
		_, r := x.QuoRem(m)
		return r
	}

	yv := []uint64{
		y.u0, y.u1, y.u2,
	}
	i := len(yv)
	for i > 0 && yv[i-1] == 0 {
		i--
	}
	yv = yv[:i]

	// TODO(eric): if x > 1 and y > 0 && mod, then use montgomery
	// or windowed exponentiation.

	z := x
	v := yv[len(yv)-1]
	s := bits.LeadingZeros64(v) + 1
	v <<= s
	w := 64 - s
	for j := 0; j < w; j++ {
		z = z.Mul(z)
		if v&mask != 0 {
			z = z.Mul(x)
		}
		if mod {
			_, z = z.QuoRem(m)
		}
		v <<= 1
	}

	for i := len(yv) - 2; i >= 0; i-- {
		v := yv[i]
		for j := 0; j < 64; j++ {
			z = z.Mul(z)
			if v&mask != 0 {
				z = z.Mul(x)
			}
			if mod {
				_, z = z.QuoRem(m)
			}
			v <<= 1
		}
	}
	return z
}

// mulPow10 returns x * 10^n.
func (x Uint256) mulPow10(n uint) (Uint256, bool) {
	switch {
	case x.IsZero():
		return Uint256{}, true
	case n == 0:
		return x, true
	case n >= 78:
		return Uint256{}, false
	default:
		return x.MulCheck(pow10Uint256(n))
	}
}

var pow10tabUint256 struct {
	values []Uint256
	once   sync.Once
}

func pow10Uint256(n uint) Uint256 {
	pow10tabUint256.once.Do(func() {
		tab := make([]Uint256, 2+78)
		tab[0] = Uint256{}
		tab[1] = U256(1)
		for i := 2; i < len(tab); i++ {
			tab[i] = tab[i-1].mul64(10)
		}
		pow10tabUint256.values = tab
	})
	return pow10tabUint256.values[n]
}

// MulCheck returns x*y and reports whether the multiplication
// oveflowed.
func (x Uint256) MulCheck(y Uint256) (Uint256, bool) {
	if x.BitLen()+y.BitLen() > 256 {
		return Uint256{}, false
	}

	var z Uint256
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, z.u0 = bits.Mul64(x.u0, d)
		c, z.u1 = mulAddWWW(x.u1, d, c)
		c, z.u2 = mulAddWWW(x.u2, d, c)
		c, z.u3 = mulAddWWW(x.u3, d, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, z.u1 = mulAddWWW(x.u0, d, z.u1)
		c, z.u2 = mulAddWWWW(x.u1, d, z.u2, c)
		c, z.u3 = mulAddWWWW(x.u2, d, z.u3, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u2 * x
	if d := y.u2; d != 0 {
		c, z.u2 = mulAddWWW(x.u0, d, z.u2)
		c, z.u3 = mulAddWWWW(x.u1, d, z.u3, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u3 * x
	if d := y.u3; d != 0 {
		c, z.u3 = mulAddWWW(x.u0, d, z.u3)
		if c != 0 {
			return Uint256{}, false
		}
	}

	return z, true
}

func (x Uint256) mulCheck64(y uint64) (Uint256, bool) {
	if y == 0 {
		return Uint256{}, true
	}
	var z Uint256
	var c uint64
	c, z.u0 = bits.Mul64(x.u0, y)
	c, z.u1 = mulAddWWW(x.u1, y, c)
	c, z.u2 = mulAddWWW(x.u2, y, c)
	c, z.u3 = mulAddWWW(x.u3, y, c)
	if c != 0 {
		return Uint256{}, false
	}
	return z, true
}

// QuoRem returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint256) QuoRem(y Uint256) (q, r Uint256) {
	if x.Cmp(y) < 0 {
		// x/y for x < y = 0.
		// x%y for x < y = x.
		return Uint256{}, x
	}

	if y.high().IsZero() {
		q, rr := x.quoRem128(y.low())
		return q, u256(rr, Uint128{})
	}

	n := uint(y.high().LeadingZeros())
	y1 := y.Lsh(n) // y1 := y<<n
	x1 := x.Rsh(1) // x1 := x>>1
	tq, _ := div128(x1.high(), x1.low(), y1.high())
	tq = tq.Rsh(127 - n) // tq >>= 127 - n
	if !tq.IsZero() {
		tq = tq.sub64(1) // tq--
	}
	q = u256(tq, Uint128{})
	ytq := y.mul128(tq) // ytq := y*tq
	r = x.Sub(ytq)      // r = x-ytq
	if r.Cmp(y) >= 0 {
		q = q.add64(1) // q++
		r = r.Sub(y)   // r -= y
	}
	return
}

// quoRem128 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint256) quoRem128(y Uint128) (q Uint256, r Uint128) {
	if x.high().Cmp(y) < 0 {
		lo, r := div128(x.high(), x.low(), y)
		return u256(lo, Uint128{}), r
	}
	hi, r := div128(Uint128{}, x.high(), y)
	lo, r := div128(r, x.low(), y)
	return u256(lo, hi), r
}

// quoRem64 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint256) quoRem64(y uint64) (q Uint256, r uint64) {
	q.u3, r = bits.Div64(0, x.u3, y)
	q.u2, r = bits.Div64(r, x.u2, y)
	q.u1, r = bits.Div64(r, x.u1, y)
	q.u0, r = bits.Div64(r, x.u0, y)
	return q, r
}

// div256 returns (q, r) such that
//
//	q = (hi, lo)/y
//	r = (hi, lo) - y*q
func div256(hi, lo, y Uint256) (q, r Uint256) {
	if y.IsZero() {
		panic("integer divide by zero")
	}
	if y.Cmp(hi) <= 0 {
		panic("integer overflow")
	}

	s := uint(y.LeadingZeros())
	y = y.Lsh(s)    // y = y<<s
	yn1 := y.high() // yn1 := y >> 128
	yn0 := y.low()  // yn0 := y & mask128

	un32 := hi.Lsh(s).Or(lo.Rsh(256 - s)) // un32 := hi<<s | lo>>(256-s)
	un10 := lo.Lsh(s)                     // un10 := lo<<s
	un1 := un10.high()                    // un1 := un10 >> 128
	un0 := un10.low()                     // un0 := un10 & mask128
	q1, rhat := un32.quoRem128(yn1)

	var c uint64 // rhat + yn1 carry

	// for q1 >= two128 || q1*yn0 > two128*rhat+un1 { ... }
	for !q1.high().IsZero() || q1.mul128(yn0).Cmp(u256(un1, rhat)) > 0 {
		q1 = q1.sub64(1)             // q1--
		rhat, c = rhat.AddCheck(yn1) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// un21 := un32*two128 + un1 - q1*y
	un21 := u256(un1, un32.low()).Sub(q1.Mul(y))
	q0, rhat := un21.quoRem128(yn1)

	// for q0 >= two128 || q0*yn0 > two128*rhat+un0 { ... }
	for !q0.high().IsZero() || q0.mul128(yn0).Cmp(u256(un0, rhat)) > 0 {
		q0 = q0.sub64(1)             // q0--
		rhat, c = rhat.AddCheck(yn1) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// q = q1*two128 + q0
	q = u256(q0.low(), q1.low())
	// r = (un21*two128 + un0 - q0*y) >> s
	r = u256(un0, un21.low()).Sub(q0.Mul(y)).Rsh(s)
	return
}

func (x Uint256) GoString() string {
	return fmt.Sprintf("[%d %d %d %d]",
		x.u0,
		x.u1,
		x.u2,
		x.u3,
	)
}

// String returns the base-10 representation of x.
func (x Uint256) String() string {
	b := make([]byte, 78)
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

// ParseUint256 returns the value of s in the given base.
func ParseUint256(s string, base int) (Uint256, error) {
	x, _, _, err := parseUint[Uint256](s, base, false)
	return x, err
}
