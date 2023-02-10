package fixed

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
)

var maxUint96 = Uint96{
	math.MaxUint64,
	math.MaxUint32,
}

// Uint96 is an unsigned, 96-bit integer.
//
// It can be compared for equality with ==.
type Uint96 struct {
	u0 uint64
	u1 uint32
}

var _ Uint[Uint96] = Uint96{}

// U96 constructs a [Uint96].
//
// The inputs are in ascending (low to high) order. For example,
// a uint64 x can be converted to a [Uint96], with
//
//	U96(x, 0, 0, ...)
//
// or more simply
//
//	U96From64(x)
func U96(u0 uint64, u1 uint32) Uint96 {
	return Uint96{u0, u1}
}

// U96From64 constructs a [Uint96] from a uint64.
func U96From64(x uint64) Uint96 {
	return Uint96{u0: x}
}

func (Uint96) max() Uint96 {
	return Uint96{
		math.MaxUint64,
		math.MaxUint32,
	}
}

func (x Uint96) uint128() Uint128 {
	return Uint128{x.u0, uint64(x.u1)}
}

//lint:ignore U1000 used by [Uint].
func (x Uint96) uint8() uint8 {
	return uint8(x.u0)
}

// digits returns the number of decimal digits required to
// represent x.
func (x Uint96) digits() int {
	if x.u1 == 0 {
		return digits(x.u0)
	}
	t := (x.BitLen() * 1233) / 4096
	if x.Cmp(pow10tab96[t]) < 0 {
		return t
	}
	return t + 1
}

func (x Uint96) words() []big.Word {
	if bits.UintSize == 32 {
		return []big.Word{
			big.Word(x.u0),
			big.Word(x.u0 >> 32),
			big.Word(x.u1),
		}
	}
	return []big.Word{
		big.Word(x.u0),
		big.Word(x.u1),
	}
}

// Bytes encodes x as a little-endian integer.
func (x Uint96) Bytes(b *[12]byte) {
	binary.LittleEndian.PutUint64(b[0:], x.u0)
	binary.LittleEndian.PutUint32(b[8:], x.u1)
}

// SetBytes sets x to the encoded little-endian integer b.
func (x *Uint96) SetBytes(b []byte) error {
	if len(b) != 12 {
		return fmt.Errorf("fixed: invalid length: %d", len(b))
	}
	x.u0 = binary.LittleEndian.Uint64(b[0:])
	x.u1 = binary.LittleEndian.Uint32(b[8:])
	return nil
}

// Size returns the width of the integer in bits.
func (Uint96) Size() int {
	return 96
}

// BitLen returns the number of bits required to represent x.
func (x Uint96) BitLen() int {
	if x.u1 != 0 {
		return 64 + bits.Len32(x.u1)
	}
	return bits.Len64(x.u0)
}

// LeadingZeros returns the number of leading zeros in x.
func (x Uint96) LeadingZeros() int {
	return 96 - x.BitLen()
}

// IsZero is shorthand for x == Uint96{}.
func (x Uint96) IsZero() bool {
	return x == Uint96{}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint96) Cmp(y Uint96) int {
	switch {
	case x == y:
		return 0
	case x.u1 < y.u1, (x.u1 == y.u1 && x.u0 < y.u0):
		return -1
	default:
		return +1
	}
}

// cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint96) cmp64(y uint64) int {
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

// Equal reports whether x == y.
//
// In general, prefer the == operator to using this method.
func (x Uint96) Equal(y Uint96) bool {
	return x == y
}

// And returns x&y.
func (x Uint96) And(y Uint96) Uint96 {
	return Uint96{x.u0 & y.u0, x.u1 & y.u1}
}

// Or returns x|y.
func (x Uint96) Or(y Uint96) Uint96 {
	return Uint96{x.u0 | y.u0, x.u1 | y.u1}
}

// orLsh64 returns x | y<<s.
//
//lint:ignore U1000 used by [Uint].
func (x Uint96) orLsh64(y uint64, s uint) Uint96 {
	return x.Or(Uint96{u0: y}.Lsh(s))
}

// Xor returns x^y.
func (x Uint96) Xor(y Uint96) Uint96 {
	return Uint96{x.u0 ^ y.u0, x.u1 ^ y.u1}
}

// Lsh returns x<<n.
func (x Uint96) Lsh(n uint) Uint96 {
	if n > 64 {
		return Uint96{0, uint32(x.u0 << (n - 64))}
	}
	return Uint96{x.u0 << n, x.u1<<n | uint32(x.u0>>(64-n))}
}

// Rsh returns x>>n.
func (x Uint96) Rsh(n uint) Uint96 {
	if n > 64 {
		return Uint96{uint64(x.u1) >> (n - 64), 0}
	}
	return Uint96{x.u0>>n | uint64(x.u1)<<(64-n), x.u1 >> n}
}

// Add returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint96) Add(y Uint96) Uint96 {
	u0, c0 := bits.Add64(x.u0, y.u0, 0)
	u1, _ := bits.Add32(x.u1, y.u1, uint32(c0))
	return Uint96{u0, u1}
}

// add64 returns x+y.
func (x Uint96) add64(y uint64) Uint96 {
	u0, c0 := bits.Add64(x.u0, y, 0)
	u1, _ := bits.Add32(x.u1, 0, uint32(c0))
	return Uint96{u0, u1}
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint96) AddCheck(y Uint96) (z Uint96, carry uint32) {
	u0, c0 := bits.Add64(x.u0, y.u0, 0)
	u1, c1 := bits.Add32(x.u1, y.u1, uint32(c0))
	return Uint96{u0, u1}, c1
}

// addCheck64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint96) addCheck64(y uint64) (z Uint96, carry uint64) {
	u0, c0 := bits.Add64(x.u0, y, 0)
	u1, c1 := bits.Add32(x.u1, 0, uint32(c0))
	return Uint96{u0, u1}, uint64(c1)
}

// Sub returns x-y.
func (x Uint96) Sub(y Uint96) Uint96 {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, _ := bits.Sub32(x.u1, y.u1, uint32(b))
	return Uint96{u0, u1}
}

func (x Uint96) sub64(y uint64) Uint96 {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, _ := bits.Sub32(x.u1, 0, uint32(b))
	return Uint96{u0, u1}
}

// SubCheck returns x-y.
//
// borrow is 1 if x+y overflows and 0 otherwise.
func (x Uint96) SubCheck(y Uint96) (z Uint96, borrow uint32) {
	u0, b0 := bits.Sub64(x.u0, y.u0, 0)
	u1, b1 := bits.Sub32(x.u1, y.u1, uint32(b0))
	return Uint96{u0, u1}, b1
}

func (x Uint96) subCheck64(y uint64) (z Uint96, borrow uint32) {
	u0, b0 := bits.Sub64(x.u0, y, 0)
	u1, b1 := bits.Sub32(x.u1, 0, uint32(b0))
	return Uint96{u0, u1}, b1
}

// Mul returns x*y.
func (x Uint96) Mul(y Uint96) Uint96 {
	u1, u0 := bits.Mul64(x.u0, y.u0)
	u1 += uint64(x.u1)*y.u0 + x.u0*uint64(y.u1)
	return Uint96{u0, uint32(u1)}
}

func (x Uint96) mul64(y uint64) Uint96 {
	u1, u0 := bits.Mul64(x.u0, y)
	return Uint96{u0, uint32(u1 + uint64(x.u1)*y)}
}

// MulCheck returns x*y and indicates whether the multiplication
// overflowed.
func (x Uint96) MulCheck(y Uint96) (Uint96, bool) {
	if x.u1 != 0 && y.u1 != 0 {
		return Uint96{}, false
	}

	var u0, u1 uint64
	var c uint64

	// y.lo * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(uint64(x.u1), d, c)
		if c != 0 || u1 > math.MaxUint32 {
			return Uint96{}, false
		}
	}

	// y.hi * x
	if d := uint64(y.u1); d != 0 {
		c, u1 = mulAddWWW(x.u0, d, uint64(u1))
		if c != 0 || u1 > math.MaxUint32 {
			return Uint96{}, false
		}
	}
	return Uint96{u0, uint32(u1)}, true
}

func (x Uint96) mulCheck64(y uint64) (Uint96, bool) {
	if y == 0 {
		return Uint96{}, false
	}
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(uint64(x.u1), y, c)
	if c != 0 || u1 > math.MaxUint32 {
		return Uint96{}, false
	}
	return Uint96{u0, uint32(u1)}, true
}

// QuoRem returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint96) QuoRem(y Uint96) (q, r Uint96) {
	if x.Cmp(y) < 0 {
		// x/y for x < y = 0.
		// x%y for x < y = x.
		return Uint96{}, x
	}
	if y.u1 == 0 {
		// Fast path for a 64-bit y.
		q, r64 := x.quoRem64(y.u0)
		return q, U96From64(r64)
	}

	// Perform 128-bit division as if the Uint96 is a Uint128
	// whose upper 32 bits are all zero.
	n := uint(bits.LeadingZeros32(y.u1))
	y1 := y.Lsh(n)
	x1 := x.Rsh(1)
	tq, _ := bits.Div64(uint64(x1.u1), x1.u0, uint64(y1.u1))
	tq >>= 63 - n
	if tq != 0 {
		tq--
	}
	q = U96From64(tq)
	ytq := y.mul64(tq) // ytq := y*tq
	r = x.Sub(ytq)     // r = x-ytq
	if r.Cmp(y) >= 0 {
		q = q.add64(1) // q--
		r = r.Sub(y)   // r -= y
	}
	return
}

// quoRem64 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint96) quoRem64(y uint64) (q Uint96, r uint64) {
	if uint64(x.u1) < y {
		lo, r := bits.Div64(uint64(x.u1), x.u0, y)
		return Uint96{lo, 0}, r
	}
	hi, r := bits.Div64(0, uint64(x.u1), y)
	lo, r := bits.Div64(r, x.u0, y)
	return Uint96{lo, uint32(hi)}, r
}

func (x Uint96) GoString() string {
	return fmt.Sprintf("[%d %d %d]",
		uint32(x.u0), x.u0>>32, x.u1)
}

// String returns the base-10 representation of x.
func (x Uint96) String() string {
	return string(x.append(nil))
}

func (x Uint96) append(dst []byte) []byte {
	if x.u1 == 0 {
		return strconv.AppendUint(dst, x.u0, 10)
	}
	b := make([]byte, maxUint96Digits)
	i := len(b)
	for x.cmp64(10) >= 0 {
		q, r := x.quoRem64(10)
		i--
		b[i] = byte(r + '0')
		x = q
	}
	i--
	b[i] = byte(x.u0 + '0')
	return append(dst, b[i:]...)
}

// ParseUint96 returns the value of s in the given base.
func ParseUint96(s string, base int) (Uint96, error) {
	x, _, _, err := parseUint[Uint96](s, base, false)
	return x, err
}

var pow10tab96 = [...]Uint96{
	{0, 0},
	{10, 0},
	{100, 0},
	{1000, 0},
	{10000, 0},
	{100000, 0},
	{1000000, 0},
	{10000000, 0},
	{100000000, 0},
	{1000000000, 0},
	{10000000000, 0},
	{100000000000, 0},
	{1000000000000, 0},
	{10000000000000, 0},
	{100000000000000, 0},
	{1000000000000000, 0},
	{10000000000000000, 0},
	{100000000000000000, 0},
	{1000000000000000000, 0},
	{10000000000000000000, 0},
	{7766279631452241920, 5},
	{3875820019684212736, 54},
	{1864712049423024128, 542},
	{200376420520689664, 5421},
	{2003764205206896640, 54210},
	{1590897978359414784, 542101},
	{15908979783594147840, 5421010},
	{11515845246265065472, 54210108},
	{4477988020393345024, 542101086},
}
