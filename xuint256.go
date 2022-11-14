//go:build ignore

package fixed

import (
	"fmt"
	"math"
	"math/bits"
)

var maxUint256 = Uint256{
	math.MaxUint64,
	math.MaxUint64,
	math.MaxUint64,
	math.MaxUint64,
}

// Uint256 is an unsigned, 256-bit integer.
//
// It can be compared for equality with ==.
type Uint256 struct {
	u0, u1, u2, u3 uint64
}

var _ Uint[Uint256] = Uint256{}

// U256 returns x as a Uint256.
func U256(x uint64) Uint256 {
	return Uint256{x, 0, 0, 0}
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

// high returns the high 128 bits in x.
func (x Uint256) high() Uint128 {
	return Uint128{x.u2, x.u3}
}

// low returns the low 128 bits in x.
func (x Uint256) low() Uint128 {
	return Uint128{x.u0, x.u1}
}

// low96 returns the low 96 bits in x.
func (x Uint256) low96() Uint96 {
	return Uint96{x.u0, uint32(x.u1)}
}

// low192 returns the low 192 bits in x.
func (x Uint256) low192() Uint192 {
	return Uint192{x.u0, x.u1, x.u2}
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
	case x == y:
		return 0
	case x.u3 < y.u3,
		x.u3 == y.u3 && x.u2 < y.u2,
		x.u3 == y.u3 && x.u2 == y.u2 && x.u1 < y.u1,
		x.u3 == y.u3 && x.u2 == y.u2 && x.u1 == y.u1 && x.u0 < y.u0:
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
func (x Uint256) cmp64(y uint64) int {
	if x.u3 != 0 || x.u2 != 0 || x.u1 != 0 {
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
func (x Uint256) And(y Uint256) Uint256 {
	return Uint256{x.u0 & y.u0, x.u1 & y.u1, x.u2 & y.u2, x.u3 & y.u3}
}

// Or returns x|y.
func (x Uint256) Or(y Uint256) Uint256 {
	return Uint256{x.u0 | y.u0, x.u1 | y.u1, x.u2 | y.u2, x.u3 | y.u3}
}

// Lsh returns x<<n.
func (x Uint256) Lsh(n uint) Uint256 {
	switch {
	case n > 192:
		return Uint256{0, 0, 0, x.u0 << (n - 192)}
	case n > 128:
		s := n - 128
		ŝ := 64 - s
		return Uint256{0, 0, x.u0 << s, x.u1<<s | x.u0>>ŝ}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint256{
			0,
			x.u0 << s,
			x.u1<<s | x.u0>>ŝ,
			x.u2<<s | x.u1>>ŝ,
		}
	default:
		s := n
		ŝ := 64 - s
		return Uint256{
			x.u0 << s,
			x.u1<<s | x.u0>>ŝ,
			x.u2<<s | x.u1>>ŝ,
			x.u3<<s | x.u2>>ŝ,
		}
	}
}

// Rsh returns x>>n.
func (x Uint256) Rsh(n uint) Uint256 {
	switch {
	case n > 192:
		s := n - 192
		return Uint256{x.u3 >> s, 0, 0, 0}
	case n > 128:
		s := n - 128
		ŝ := 64 - s
		return Uint256{
			x.u2>>s | x.u3<<ŝ,
			x.u3 >> s,
			0,
			0,
		}
	case n > 64:
		s := n - 64
		ŝ := 64 - s
		return Uint256{
			x.u1>>s | x.u2<<ŝ,
			x.u2>>s | x.u3<<ŝ,
			x.u3 >> s,
			0,
		}
	default:
		s := n
		ŝ := 64 - s
		return Uint256{
			x.u0>>s | x.u1<<ŝ,
			x.u1>>s | x.u2<<ŝ,
			x.u2>>s | x.u3<<ŝ,
			x.u3 >> s,
		}
	}
}

// Add returns x+y.
func (x Uint256) Add(y Uint256) Uint256 {
	u0, c := bits.Add64(x.u0, y.u0, 0)
	u1, c := bits.Add64(x.u1, y.u1, c)
	u2, c := bits.Add64(x.u2, y.u2, c)
	u3, _ := bits.Add64(x.u3, y.u3, c)
	return Uint256{u0, u1, u2, u3}
}

// add64 returns x+y.
func (x Uint256) add64(y uint64) Uint256 {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, c := bits.Add64(x.u1, 0, c)
	u2, c := bits.Add64(x.u2, 0, c)
	u3, _ := bits.Add64(x.u3, 0, c)
	return Uint256{u0, u1, u2, u3}
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint256) AddCheck(y Uint256) (z Uint256, carry uint64) {
	u0, c := bits.Add64(x.u0, y.u0, 0)
	u1, c := bits.Add64(x.u1, y.u1, c)
	u2, c := bits.Add64(x.u2, y.u2, c)
	u3, c := bits.Add64(x.u3, y.u3, c)
	return Uint256{u0, u1, u2, u3}, c
}

// add64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x Uint256) addCheck64(y uint64) (z Uint256, carry uint64) {
	u0, c := bits.Add64(x.u0, y, 0)
	u1, c := bits.Add64(x.u1, 0, c)
	u2, c := bits.Add64(x.u2, 0, c)
	u3, c := bits.Add64(x.u3, 0, c)
	return Uint256{u0, u1, u2, u3}, c
}

// Sub returns x-y.
func (x Uint256) Sub(y Uint256) Uint256 {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, b := bits.Sub64(x.u1, y.u1, b)
	u2, b := bits.Sub64(x.u2, y.u2, b)
	u3, _ := bits.Sub64(x.u3, y.u3, b)
	return Uint256{u0, u1, u2, u3}
}

// sub64 returns x-y.
func (x Uint256) sub64(y uint64) Uint256 {
	u0, b := bits.Sub64(x.u0, y, 0)
	u1, b := bits.Sub64(x.u1, 0, b)
	u2, b := bits.Sub64(x.u2, 0, b)
	u3, _ := bits.Sub64(x.u3, 0, b)
	return Uint256{u0, u1, u2, u3}
}

// SubCheck returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x Uint256) SubCheck(y Uint256) (z Uint256, borrow uint64) {
	u0, b := bits.Sub64(x.u0, y.u0, 0)
	u1, b := bits.Sub64(x.u1, y.u1, b)
	u2, b := bits.Sub64(x.u2, y.u2, b)
	u3, b := bits.Sub64(x.u3, y.u3, b)
	return Uint256{u0, u1, u2, u3}, b
}

// Mul returns x*y.
func (x Uint256) Mul(y Uint256) Uint256 {
	var u0, u1, u2, u3 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		c, u2 = mulAddWWW(x.u2, d, c)
		u3 = x.u3*d + c
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		c, u2 = mulAddWWWW(x.u1, d, u2, c)
		u3 += x.u2*d + c
	}

	// y.u2 * x
	if d := y.u2; d != 0 {
		c, u2 = mulAddWWW(x.u0, d, u2)
		u3 += x.u1*d + c
	}

	// y.u3 * x
	u3 += x.u0 * y.u3

	return Uint256{u0, u1, u2, u3}
}

// MulCheck returns x*y and reports whether the multiplication
// oveflowed.
func (x Uint256) MulCheck(y Uint256) (Uint256, bool) {
	if x.BitLen()+y.BitLen() > 256 {
		return Uint256{}, false
	}

	var u0, u1, u2, u3 uint64
	var c uint64

	// y.u0 * x
	if d := y.u0; d != 0 {
		c, u0 = bits.Mul64(x.u0, d)
		c, u1 = mulAddWWW(x.u1, d, c)
		c, u2 = mulAddWWW(x.u2, d, c)
		c, u3 = mulAddWWW(x.u3, d, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u1 * x
	if d := y.u1; d != 0 {
		c, u1 = mulAddWWW(x.u0, d, u1)
		c, u2 = mulAddWWWW(x.u1, d, u2, c)
		c, u3 = mulAddWWWW(x.u2, d, u3, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u2 * x
	if d := y.u2; d != 0 {
		c, u2 = mulAddWWW(x.u0, d, u2)
		c, u3 = mulAddWWWW(x.u1, d, u3, c)
		if c != 0 {
			return Uint256{}, false
		}
	}

	// y.u3 * x
	if d := y.u3; d != 0 {
		c, u3 = mulAddWWW(x.u0, d, u3)
		if c != 0 {
			return Uint256{}, false
		}
	}
	return Uint256{u0, u1, u2, u3}, true
}

func (x Uint256) mulCheck64(y uint64) (Uint256, bool) {
	if y == 0 {
		return Uint256{}, true
	}
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(x.u1, y, c)
	c, u2 := mulAddWWW(x.u2, y, c)
	c, u3 := mulAddWWW(x.u3, y, c)
	if c != 0 {
		return Uint256{}, false
	}
	return Uint256{u0, u1, u2, u3}, true
}

func (x Uint256) mul64(y uint64) Uint256 {
	c, u0 := bits.Mul64(x.u0, y)
	c, u1 := mulAddWWW(x.u1, y, c)
	c, u2 := mulAddWWW(x.u2, y, c)
	u3 := x.u3*y + c
	return Uint256{u0, u1, u2, u3}
}

func (x Uint256) mul128(y Uint128) Uint256 {
	return x.Mul(y.uint256())
}

// mulPow10 returns x * 10^n.
func (x Uint256) mulPow10(n uint) (Uint256, bool) {
	switch {
	case x.IsZero():
		return Uint256{}, true
	case n == 0:
		return x, true
	case n >= uint(len(pow10tab256)):
		return Uint256{}, false
	default:
		return x.MulCheck(pow10tab256[n])
	}
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
		if y.u1 == 0 {
			// Fast path for a 64-bit y.
			q, r64 := x.quoRem64(y.u0)
			return q, U256(r64)
		}
		// Fast path for a 128-bit y.
		q, r128 := x.quoRem128(y.low())
		return q, Uint256{r128.u0, r128.u1, 0, 0}
	}

	n := uint(y.high().LeadingZeros())
	y1 := y.Lsh(n) // y1 := y<<n
	x1 := x.Rsh(1) // x1 := x>>1
	tq, _ := div128(x1.high(), x1.low(), y1.high())
	tq = tq.Rsh(127 - n) // tq >>= 127 - n
	if !tq.IsZero() {
		tq = tq.sub64(1) // tq--
	}
	q = tq.uint256()
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
func (x Uint256) quoRem64(y uint64) (q Uint256, r uint64) {
	q.u3, r = bits.Div64(0, x.u3, y)
	q.u2, r = bits.Div64(r, x.u2, y)
	q.u1, r = bits.Div64(r, x.u1, y)
	q.u0, r = bits.Div64(r, x.u0, y)
	return
}

// quoRem128 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x Uint256) quoRem128(y Uint128) (q Uint256, r Uint128) {
	if x.high().Cmp(y) < 0 {
		lo, r := div128(x.high(), x.low(), y)
		return Uint256{lo.u0, lo.u1, 0, 0}, r
	}
	hi, r := div128(Uint128{}, x.high(), y)
	lo, r := div128(r, x.low(), y)
	return Uint256{lo.u0, lo.u1, hi.u0, hi.u1}, r
}

// div256 returns (q, r) such that
//
//  q = (hi, lo)/y
//  r = (hi, lo) - y*q
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
	return fmt.Sprintf("[%d %d %d %d]", x.u0, x.u1, x.u2, x.u3)
}

// String returns the base-10 representation of x.
func (x Uint256) String() string {
	if x.u3 == 0 {
		return x.low192().String()
	}
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

var pow10tab256 = [...]Uint256{
	{0, 0, 0, 0},
	{10, 0, 0, 0},
	{100, 0, 0, 0},
	{1000, 0, 0, 0},
	{10000, 0, 0, 0},
	{100000, 0, 0, 0},
	{1000000, 0, 0, 0},
	{10000000, 0, 0, 0},
	{100000000, 0, 0, 0},
	{1000000000, 0, 0, 0},
	{10000000000, 0, 0, 0},
	{100000000000, 0, 0, 0},
	{1000000000000, 0, 0, 0},
	{10000000000000, 0, 0, 0},
	{100000000000000, 0, 0, 0},
	{1000000000000000, 0, 0, 0},
	{10000000000000000, 0, 0, 0},
	{100000000000000000, 0, 0, 0},
	{1000000000000000000, 0, 0, 0},
	{10000000000000000000, 0, 0, 0},
	{7766279631452241920, 5, 0, 0},
	{3875820019684212736, 54, 0, 0},
	{1864712049423024128, 542, 0, 0},
	{200376420520689664, 5421, 0, 0},
	{2003764205206896640, 54210, 0, 0},
	{1590897978359414784, 542101, 0, 0},
	{15908979783594147840, 5421010, 0, 0},
	{11515845246265065472, 54210108, 0, 0},
	{4477988020393345024, 542101086, 0, 0},
	{7886392056514347008, 5421010862, 0, 0},
	{5076944270305263616, 54210108624, 0, 0},
	{13875954555633532928, 542101086242, 0, 0},
	{9632337040368467968, 5421010862427, 0, 0},
	{4089650035136921600, 54210108624275, 0, 0},
	{4003012203950112768, 542101086242752, 0, 0},
	{3136633892082024448, 5421010862427522, 0, 0},
	{12919594847110692864, 54210108624275221, 0, 0},
	{68739955140067328, 542101086242752217, 0, 0},
	{687399551400673280, 5421010862427522170, 0, 0},
	{6873995514006732800, 17316620476856118468, 2, 0},
	{13399722918938673152, 7145508105175220139, 29, 0},
	{4870020673419870208, 16114848830623546549, 293, 0},
	{11806718586779598848, 13574535716559052564, 2938, 0},
	{7386721425538678784, 6618148649623664334, 29387, 0},
	{80237960548581376, 10841254275107988496, 293873, 0},
	{802379605485813760, 16178822382532126880, 2938735, 0},
	{8023796054858137600, 14214271235644855872, 29387358, 0},
	{6450984253743169536, 13015503840481697412, 293873587, 0},
	{9169610316303040512, 1027829888850112811, 2938735877, 0},
	{17909126868192198656, 10278298888501128114, 29387358770, 0},
	{13070572018536022016, 10549268516463523069, 293873587705, 0},
	{1578511669393358848, 13258964796087472617, 2938735877055, 0},
	{15785116693933588480, 3462439444907864858, 29387358770557, 0},
	{10277214349659471872, 16177650375369096972, 293873587705571, 0},
	{10538423128046960640, 14202551164014556797, 2938735877055718, 0},
	{13150510911921848320, 12898303124178706663, 29387358770557187, 0},
	{2377900603251621888, 18302566799529756941, 293873587705571876, 0},
	{5332261958806667264, 17004971331911604867, 2938735877055718769, 0},
	{16429131440647569408, 4029016655730084128, 10940614696847636083, 1},
	{16717361816799281152, 3396678409881738056, 17172426599928602752, 15},
	{1152921504606846976, 15520040025107828953, 5703569335900062977, 159},
	{11529215046068469760, 7626447661401876602, 1695461137871974930, 1593},
	{4611686018427387904, 2477500319180559562, 16954611378719749304, 15930},
	{9223372036854775808, 6328259118096044006, 3525417123811528497, 159309},
	{0, 7942358959831785217, 16807427164405733357, 1593091},
	{0, 5636613303479645706, 2053574980671369030, 15930919},
	{0, 1025900813667802212, 2089005733004138687, 159309191},
	{0, 10259008136678022120, 2443313256331835254, 1593091911},
	{0, 10356360998232463120, 5986388489608800929, 15930919111},
	{0, 11329889613776873120, 4523652674959354447, 159309191113},
	{0, 2618431695511421504, 8343038602174441244, 1593091911132},
	{0, 7737572881404663424, 9643409726906205977, 15930919111324},
	{0, 3588752519208427776, 4200376900514301694, 159309191113245},
	{0, 17440781118374726144, 5110280857723913709, 1593091911132452},
	{0, 8387114520361296896, 14209320429820033867, 15930919111324522},
	{0, 10084168908774762496, 12965995782233477362, 159309191113245227},
	{0, 8607968719199866880, 532749306367912313, 1593091911132452277},
	{0, 12292710897160462336, 5327493063679123134, 15930919111324522770},
}
