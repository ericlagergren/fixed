package fixed

import (
	"fmt"
	"math"
	"math/bits"
)

var maxUint192 = Uint192{
	math.MaxUint64,
	math.MaxUint64,
	math.MaxUint64,
}

// Uint192 is an unsigned, 192-bit integer.
//
// It can be compared for equality with ==.
type Uint192 struct {
	u0, u1, u2 uint64
}

var _ Uint[Uint192] = Uint192{}

// U192 returns x as a Uint192.
func U192(x uint64) Uint192 {
	return Uint192{x, 0, 0}
}

func (Uint192) max() Uint192 {
	return Uint192{
		math.MaxUint64,
		math.MaxUint64,
		math.MaxUint64,
	}
}

// high returns the high 96 bits in x.
func (x Uint192) high() Uint96 {
	return Uint96{x.u1>>32 | x.u2<<32, uint32(x.u2 >> 32)}
}

// low128 returns the low 128 bits in x.
func (x Uint192) low128() Uint128 {
	return Uint128{x.u0, x.u1}
}

// hi128 returns the high 128 bits in x.
//
// Since x is 192 bits, the high 64 bits in the result are always
// zero.
func (x Uint192) hi128() Uint128 {
	return Uint128{x.u2, 0}
}

// BitLen returns the number of bits required to represent x.
func (x Uint192) BitLen() int {
	switch {
	case x.u2 != 0:
		return 128 + bits.Len64(x.u2)
	case x.u1 != 0:
		return 64 + bits.Len64(x.u1)
	default:
		return bits.Len64(x.u0)
	}
}

// LeadingZeros returns the number of leading zeros in x.
func (x Uint192) LeadingZeros() int {
	return 192 - x.BitLen()
}

// IsZero is shorthand for x == Uint192{}.
func (x Uint192) IsZero() bool {
	return x == Uint192{}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x Uint192) Cmp(y Uint192) int {
	switch {
	case x == y:
		return 0
	case x.u2 < y.u2,
		x.u2 == y.u2 && x.u1 < y.u1,
		x.u2 == y.u2 && x.u1 == y.u1 && x.u0 < y.u0:
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
func (x Uint192) cmp64(y uint64) int {
	if x.u2 != 0 || x.u1 != 0 {
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
func (x Uint192) And(y Uint192) Uint192 {
	return Uint192{x.u0 & y.u0, x.u1 & y.u1, x.u2 & y.u2}
}

// Or returns x|y.
func (x Uint192) Or(y Uint192) Uint192 {
	return Uint192{x.u0 | y.u0, x.u1 | y.u1, x.u2 | y.u2}
}

// Xor returns x^y.
func (x Uint192) Xor(y Uint192) Uint192 {
	return Uint192{x.u0 ^ y.u0, x.u1 ^ y.u1, x.u2 ^ y.u2}
}

// Lsh returns x<<n.
func (x Uint192) Lsh(n uint) Uint192 {
	switch {
	case n > 128:
		return Uint192{0, 0, x.u0 << (n - 128)}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint192{
			0,
			x.u0 << s,
			x.u1<<s | x.u0>>ŝ,
		}
	default:
		s := n
		ŝ := 64 - s
		return Uint192{
			x.u0 << s,
			x.u1<<s | x.u0>>ŝ,
			x.u2<<s | x.u1>>ŝ,
		}
	}
}

// Rsh returns x>>n.
func (x Uint192) Rsh(n uint) Uint192 {
	switch {
	case n > 128:
		return Uint192{x.u2 >> (n - 128), 0, 0}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint192{x.u1>>s | x.u2<<ŝ, x.u2 >> s, 0}
	default:
		s := n
		ŝ := 64 - s
		return Uint192{
			x.u0>>s | x.u1<<ŝ,
			x.u1>>s | x.u2<<ŝ,
			x.u2 >> s,
		}
	}
}

// Add returns x+y.
func (x Uint192) Add(y Uint192) Uint192 {
	u0, c := bits.Add64(x.u0, y.u0, 0)
	u1, c := bits.Add64(x.u1, y.u1, c)
	u2, _ := bits.Add64(x.u2, y.u2, c)
	return Uint192{u0, u1, u2}
}

// add64 returns x+y.
func (x Uint192) add64(y uint64) Uint192 {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, c := bits.Add64(x.u1, 0, c)
	u2, _ := bits.Add64(x.u2, 0, c)
	return Uint192{u0, u1, u2}
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint192) AddCheck(y Uint192) (z Uint192, carry uint64) {
	u0, c := bits.Add64(x.u0, y.u0, 0)
	u1, c := bits.Add64(x.u1, y.u1, c)
	u2, c := bits.Add64(x.u2, y.u2, c)
	return Uint192{u0, u1, u2}, c
}

// addCheck64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint192) addCheck64(y uint64) (z Uint192, carry uint64) {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, c := bits.Add64(x.u1, 0, c)
	u2, c := bits.Add64(x.u2, 0, c)
	return Uint192{u0, u1, u2}, c
}

// Sub returns x-y.
func (x Uint192) Sub(y Uint192) Uint192 {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, b := bits.Sub64(x.u1, y.u1, b)
	u2, _ := bits.Sub64(x.u2, y.u2, b)
	return Uint192{u0, u1, u2}
}

func (x Uint192) sub64(y uint64) Uint192 {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, b := bits.Sub64(x.u1, 0, b)
	u2, _ := bits.Sub64(x.u2, 0, b)
	return Uint192{u0, u1, u2}
}

// SubCheck returns x-y.
//
// borrow is 1 if x+y overflows and 0 otherwise.
func (x Uint192) SubCheck(y Uint192) (z Uint192, borrow uint64) {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, b := bits.Sub64(x.u1, y.u1, b)
	u2, b := bits.Sub64(x.u2, y.u2, b)
	return Uint192{u0, u1, u2}, b
}

func (x Uint192) subCheck64(y uint64) (z Uint192, borrow uint64) {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, b := bits.Sub64(x.u1, 0, b)
	u2, b := bits.Sub64(x.u2, 0, b)
	return Uint192{u0, u1, u2}, b
}

// Mul returns x*y.
func (x Uint192) Mul(y Uint192) Uint192 {
	var u0, u1, u2 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		u2 = x.u2*d + c
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		u2 += x.u1*d + c
	}

	// y.u2 * x
	u2 += x.u0 * y.u2

	return Uint192{u0, u1, u2}
}

func (x Uint192) mul64(y uint64) Uint192 {
	if y == 0 {
		return Uint192{}
	}
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(x.u1, y, c)
	u2 := x.u2*y + c
	return Uint192{u0, u1, u2}
}

// mul128 returns x*y.
func (x Uint192) mul128(y Uint128) Uint192 {
	var u0, u1, u2 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		u2 = x.u2*d + c
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		u2 += x.u1*d + c
	}
	return Uint192{u0, u1, u2}
}

// MulCheck returns x*y and indicates whether the multiplication
// overflowed.
func (x Uint192) MulCheck(y Uint192) (Uint192, bool) {
	if x.BitLen()+y.BitLen() > 192 {
		return Uint192{}, false
	}

	var u0, u1, u2 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		c, u2 = mulAddWWW(x.u2, d, c)
		if c != 0 {
			return Uint192{}, false
		}
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		c, u2 = mulAddWWWW(x.u1, d, u2, c)
		if c != 0 {
			return Uint192{}, false
		}
	}

	// y.u2 * x
	if d := y.u2; d != 0 {
		c, u2 = mulAddWWW(x.u0, d, u2)
		if c != 0 {
			return Uint192{}, false
		}
	}
	return Uint192{u0, u1, u2}, true
}

func (x Uint192) mulCheck64(y uint64) (Uint192, bool) {
	// TODO(eric): make this inlinable.
	if y == 0 {
		return Uint192{}, true
	}
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(x.u1, y, c)
	c, u2 := mulAddWWW(x.u2, y, c)
	if c != 0 {
		return Uint192{}, false
	}
	return Uint192{u0, u1, u2}, true
}

// QuoRem returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint192) QuoRem(y Uint192) (q, r Uint192) {
	if x.Cmp(y) < 0 {
		// x/y for x < y = 0.
		// x%y for x < y = x.
		return Uint192{}, x
	}
	if y.u2 == 0 {
		if y.u1 == 0 {
			// Fast path for a 64-bit y.
			q, r64 := x.quoRem64(y.u0)
			return q, U192(r64)
		}
		// Fast path for a 128-bit y.
		q, r128 := x.quoRem128(y.low128())
		return q, r128.uint192()
	}

	n := uint(y.high().LeadingZeros())
	y1 := y.Lsh(n) // y1 := y<<n
	x1 := x.Rsh(1) // x1 := x>>1
	tq, _ := div128(x1.hi128(), x1.low128(), y1.hi128())
	tq = tq.Rsh(127 - n) // tq >>= 127 - n
	if !tq.IsZero() {
		tq = tq.sub64(1) // tq--
	}
	q = tq.uint192()
	ytq := y.mul128(tq) // ytq := y*tq
	r = x.Sub(ytq)      // r = x-ytq
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
func (x Uint192) quoRem64(y uint64) (q Uint192, r uint64) {
	u2, r := bits.Div64(0, x.u2, y)
	u1, r := bits.Div64(r, x.u1, y)
	u0, r := bits.Div64(r, x.u0, y)
	return Uint192{u0, u1, u2}, r
}

// quoRem96 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint192) quoRem96(y Uint96) (q Uint192, r Uint96) {
	q, rr := x.quoRem128(y.uint128())
	return q, rr.uint96()
}

// quoRem128 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint192) quoRem128(y Uint128) (q Uint192, r Uint128) {
	if x.hi128().Cmp(y) < 0 {
		lo, r := div128(x.hi128(), x.low128(), y)
		return lo.uint192(), r
	}
	hi, r := div128(Uint128{}, x.hi128(), y)
	lo, r := div128(r, x.low128(), y)
	return Uint192{lo.u0, lo.u1, hi.u0}, r
}

func (x Uint192) GoString() string {
	return fmt.Sprintf("[%d %d %d]", x.u0, x.u1, x.u2)
}

// String returns the base-10 representation of x.
func (x Uint192) String() string {
	if x.u2 == 0 {
		return x.low128().String()
	}
	b := make([]byte, 58)
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

// ParseUint192 returns the value of s in the given base.
func ParseUint192(s string, base int) (Uint192, error) {
	x, _, _, err := parseUint[Uint192](s, base, false)
	return x, err
}

var pow10tab192 = [...]Uint192{
	{0, 0, 0},
	{10, 0, 0},
	{100, 0, 0},
	{1000, 0, 0},
	{10000, 0, 0},
	{100000, 0, 0},
	{1000000, 0, 0},
	{10000000, 0, 0},
	{100000000, 0, 0},
	{1000000000, 0, 0},
	{10000000000, 0, 0},
	{100000000000, 0, 0},
	{1000000000000, 0, 0},
	{10000000000000, 0, 0},
	{100000000000000, 0, 0},
	{1000000000000000, 0, 0},
	{10000000000000000, 0, 0},
	{100000000000000000, 0, 0},
	{1000000000000000000, 0, 0},
	{10000000000000000000, 0, 0},
	{7766279631452241920, 5, 0},
	{3875820019684212736, 54, 0},
	{1864712049423024128, 542, 0},
	{200376420520689664, 5421, 0},
	{2003764205206896640, 54210, 0},
	{1590897978359414784, 542101, 0},
	{15908979783594147840, 5421010, 0},
	{11515845246265065472, 54210108, 0},
	{4477988020393345024, 542101086, 0},
	{7886392056514347008, 5421010862, 0},
	{5076944270305263616, 54210108624, 0},
	{13875954555633532928, 542101086242, 0},
	{9632337040368467968, 5421010862427, 0},
	{4089650035136921600, 54210108624275, 0},
	{4003012203950112768, 542101086242752, 0},
	{3136633892082024448, 5421010862427522, 0},
	{12919594847110692864, 54210108624275221, 0},
	{68739955140067328, 542101086242752217, 0},
	{687399551400673280, 5421010862427522170, 0},
	{6873995514006732800, 17316620476856118468, 2},
	{13399722918938673152, 7145508105175220139, 29},
	{4870020673419870208, 16114848830623546549, 293},
	{11806718586779598848, 13574535716559052564, 2938},
	{7386721425538678784, 6618148649623664334, 29387},
	{80237960548581376, 10841254275107988496, 293873},
	{802379605485813760, 16178822382532126880, 2938735},
	{8023796054858137600, 14214271235644855872, 29387358},
	{6450984253743169536, 13015503840481697412, 293873587},
	{9169610316303040512, 1027829888850112811, 2938735877},
	{17909126868192198656, 10278298888501128114, 29387358770},
	{13070572018536022016, 10549268516463523069, 293873587705},
	{1578511669393358848, 13258964796087472617, 2938735877055},
	{15785116693933588480, 3462439444907864858, 29387358770557},
	{10277214349659471872, 16177650375369096972, 293873587705571},
	{10538423128046960640, 14202551164014556797, 2938735877055718},
	{13150510911921848320, 12898303124178706663, 29387358770557187},
	{2377900603251621888, 18302566799529756941, 293873587705571876},
	{5332261958806667264, 17004971331911604867, 2938735877055718769},
}
