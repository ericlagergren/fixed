package fixed

import (
	"math/big"
	"testing"

	"golang.org/x/exp/rand"
)

var (
	big96mask    = bigMask(96)
	bigMaxUint96 = Uint96{}.max().big()
)

func randUint96() Uint96 {
	var x Uint96
	if randBool() {
		x.u0 = rand.Uint64()
	}
	if randBool() {
		x.u1 = rand.Uint32()
	}
	return x
}

func (x Uint96) big() *big.Int {
	var v big.Int
	v.SetBits(x.words())
	return &v
}

func TestUint96Bytes(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := randUint96()
		var b [12]byte
		x.Bytes(&b)
		var y Uint96
		if err := y.SetBytes(b[:]); err != nil {
			t.Fatal(err)
		}
		if x != y {
			t.Fatalf("got %x, expected %x", y, x)
		}
	}
}

func TestUint96Digits(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		got := x.digits()
		want := len(x.String())
		if want != got {
			t.Fatalf("%s: expected %d, got %d", x, want, got)
		}
	}
}

func TestUint96Cmp(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint96()

		got := x.Cmp(y)
		want := x.big().Cmp(y.big())
		if got != want {
			t.Fatalf("cmp(%d, %d): expected %d, got %d",
				x.big(), y.big(), want, got)
		}
	}
}

func TestUint96Add(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint96()

		z, c := x.AddCheck(y)

		want := new(big.Int).Add(x.big(), y.big())
		if carry := want.Cmp(bigMaxUint96) > 0; carry != (c == 1) {
			t.Fatalf("%d + %d: expected %t, got %t",
				x.big(), y.big(), carry, c == 1)
		}
		want.And(want, big96mask)

		if c == 0 && x.Add(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), y.big(), x.Add(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d + %d: expected %d, got %d",
				x.big(), y.big(), want, got)
		}
	}
}

func TestUint96Add64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := randUint96()
		y := randUint64()

		z, c := x.addCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Add(x.big(), ybig)
		if carry := want.Cmp(bigMaxUint96) > 0; carry != (c == 1) {
			t.Fatalf("%d + %d: expected %t, got %t",
				x.big(), ybig, carry, c == 1)
		}
		want.And(want, big96mask)

		if c == 0 && x.add64(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), ybig, x.add64(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d + %d: expected %d, got %d",
				x.big(), ybig, want, got)
		}
	}
}

func TestUint96Sub(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint96()

		z, b := x.SubCheck(y)

		want := new(big.Int).Sub(x.big(), y.big())
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%d - %d: expected %t, got %t",
				x.big(), y.big(), borrow, b == 1)
		}
		want.And(want, big96mask)

		if b == 0 && x.Sub(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), y.big(), x.Sub(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d - %d: expected %d, got %d",
				x.big(), y.big(), want, got)
		}
	}
}

func TestUint96Sub64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := randUint96()
		y := randUint64()

		z, b := x.subCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Sub(x.big(), ybig)
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%d - %d: expected %t, got %t",
				x.big(), ybig, borrow, b == 1)
		}
		want.And(want, big96mask)

		if b == 0 && x.sub64(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), ybig, x.sub64(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d - %d: expected %d, got %d",
				x.big(), ybig, want, got)
		}
	}
}

func TestUint96Mul(t *testing.T) {
	for i := 0; i < 500_000; i++ {
		x := randUint96()
		y := randUint96()

		z, ok := x.MulCheck(y)

		want := new(big.Int).Mul(x.big(), y.big())
		if good := want.Cmp(bigMaxUint96) <= 0; good != ok {
			t.Fatalf("%d: %d * %d: expected %t, got %t",
				i, x.big(), y.big(), good, ok)
		}
		want.And(want, big96mask)

		if ok && x.Mul(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), y.big(), x.Mul(y), z)
		}
		z = x.Mul(y)
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d: %d * %d: expected %d, got %d",
				i, x.big(), y.big(), want, got)
		}
	}
}

func TestUint96Mul64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint64()

		z, ok := x.mulCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Mul(x.big(), ybig)
		if (want.Cmp(bigMaxUint96) <= 0) != ok {
			t.Fatalf("%d: %d * %d: expected %t",
				i, x.big(), ybig, !ok)
		}
		want.And(want, big96mask)

		if ok && x.mul64(y) != z {
			t.Fatalf("%d: %d * %d: %d != %d",
				i, x.big(), ybig, x.mul64(y), z)
		}
		z = x.mul64(y)
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d: %d * %d: expected %d, got %d",
				i, x.big(), ybig, want, got)
		}
	}
}

func TestUint96QuoRem(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint96()
		if y == (Uint96{}) {
			y = U96From64(1)
		}

		q, r := x.QuoRem(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big96mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%d: %d / %d expected quotient of %d, got %d",
				i, x.big(), y.big(), wantq, got)
		}
		if got := r.big(); got.Cmp(wantr) != 0 {
			t.Fatalf("%d: %d / %d expected remainder of %d, got %d",
				i, x.big(), y.big(), wantr, got)
		}
	}
}

func TestUint96QuoRem64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint96()
		y := randUint64()
		if y == 0 {
			y = 1
		}

		q, r := x.quoRem64(y)

		ybig := new(big.Int).SetUint64(y)
		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), ybig, wantr)
		wantq.And(wantq, big96mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%d / %d expected quotient of %d, got %d",
				x.big(), ybig, wantq, got)
		}
		if got := new(big.Int).SetUint64(r); got.Cmp(wantr) != 0 {
			t.Fatalf("%d / %d expected remainder of %d, got %d",
				x.big(), ybig, wantr, got)
		}
	}
}

func TestUint96String(t *testing.T) {
	test := func(x Uint96) {
		want := x.big().String()
		got := x.String()
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
	test(Uint96{})  // min
	test(maxUint96) // max
	for i := 0; i < 100_000; i++ {
		test(randUint96())
	}
}

func TestParseUint96(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		want := randUint96()
		b := want.big()
		for base := 2; base <= 36; base++ {
			s := b.Text(base)
			got, err := ParseUint96(s, base)
			if err != nil {
				t.Fatalf("%q in base %d: unexpected error: %v", s, base, err)
			}
			if got != want {
				t.Fatalf("%q in base %d: expected %#v, got %#v",
					s, base, want, got)
			}
		}
	}
}

func BenchmarkUint96Mul(b *testing.B) {
	s := make([]Uint96, 1000)
	for i := range s {
		s[i] = randUint96()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%len(s)]
		y := s[(i+1)%len(s)]
		sink.Uint96 = x.Mul(y)
	}
}

func BenchmarkUint96QuoRem64(b *testing.B) {
	b.Run("obvious", func(b *testing.B) {
		benchmarkUint96QuoRem64(b, Uint96.quoRem64)
	})
	b.Run("reciprocal", func(b *testing.B) {
		benchmarkUint96QuoRem64(b, Uint96.quoRem64Reciprocal)
	})
}

func (x Uint96) quoRem64Reciprocal(y uint64) (q Uint96, r uint64) {
	rec := reciprocal(y)
	hi, r := divWW(0, uint64(x.u1), y, rec)
	lo, r := divWW(r, x.u0, y, rec)
	return Uint96{lo, uint32(hi)}, r
}

func benchmarkUint96QuoRem64(b *testing.B, fn func(x Uint96, y uint64) (q Uint96, r uint64)) {
	for i := 0; i < b.N; i++ {
		sink.Uint96, sink.uint64 = fn(U96From64(uint64(i+2)), uint64(i+1))
	}
}
