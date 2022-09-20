package fixed

import (
	"math/big"
	"math/bits"
	"testing"

	"golang.org/x/exp/rand"
)

var (
	big192mask    = bigMask(192)
	bigMaxUint192 = Uint192{}.max().big()
)

func randUint192() Uint192 {
	var x Uint192
	if randBool() {
		x.u0 = rand.Uint64()
	}
	if randBool() {
		x.u1 = rand.Uint64()
	}
	if randBool() {
		x.u2 = rand.Uint64()
	}
	return x
}

func (x Uint192) big() *big.Int {
	var v big.Int
	if bits.UintSize == 32 {
		v.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u0 >> 32),
			big.Word(x.u1),
			big.Word(x.u1 >> 32),
			big.Word(x.u2),
			big.Word(x.u2 >> 32),
		})
	} else {
		v.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u1),
			big.Word(x.u2),
		})
	}
	return &v
}

func TestUint192Cmp(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint192()

		got := x.Cmp(y)
		want := x.big().Cmp(y.big())
		if got != want {
			t.Fatalf("cmp(%d, %d): expected %d, got %d",
				x.big(), y.big(), want, got)
		}
	}
}

func TestUint192Add(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint192()

		z, c := x.AddCheck(y)

		want := new(big.Int).Add(x.big(), y.big())
		if carry := want.Cmp(bigMaxUint192) > 0; carry != (c == 1) {
			t.Fatalf("%d + %d: expected %t, got %t",
				x.big(), y.big(), carry, c == 1)
		}
		want.And(want, big192mask)

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

func TestUint192Add64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := randUint192()
		y := randUint64()

		z, c := x.addCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Add(x.big(), ybig)
		if carry := want.Cmp(bigMaxUint192) > 0; carry != (c == 1) {
			t.Fatalf("%d + %d: expected %t, got %t",
				x.big(), ybig, carry, c == 1)
		}
		want.And(want, big192mask)

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

func TestUint192Sub(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint192()

		z, b := x.SubCheck(y)

		want := new(big.Int).Sub(x.big(), y.big())
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%d - %d: expected %t, got %t",
				x.big(), y.big(), borrow, b == 1)
		}
		want.And(want, big192mask)

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

func TestUint192Sub64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := randUint192()
		y := randUint64()

		z, b := x.subCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Sub(x.big(), ybig)
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%d - %d: expected %t, got %t",
				x.big(), ybig, borrow, b == 1)
		}
		want.And(want, big192mask)

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

func TestUint192Mul(t *testing.T) {
	for i := 0; i < 750_000; i++ {
		x := randUint192()
		y := randUint192()

		z, ok := x.MulCheck(y)

		want := new(big.Int).Mul(x.big(), y.big())
		if good := want.Cmp(bigMaxUint192) <= 0; good != ok {
			t.Fatalf("%d: %d * %d: expected %t, got %t",
				i, x.big(), y.big(), good, ok)
		}
		want.And(want, big192mask)

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

func TestUint192Mul64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint64()

		z, ok := x.mulCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Mul(x.big(), ybig)
		if (want.Cmp(bigMaxUint192) <= 0) != ok {
			t.Fatalf("%d: %d * %d: expected %t",
				i, x.big(), ybig, !ok)
		}
		want.And(want, big192mask)

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

func TestUint192QuoRem(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint192()
		if y == (Uint192{}) {
			y = U192(1)
		}

		q, r := x.QuoRem(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big192mask)

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

func TestUint192QuoRemHalf(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint96()
		if y.IsZero() {
			y = U96(1)
		}

		q, r := x.quoRem96(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big192mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%d / %d expected quotient of %d, got %d",
				x.big(), y.big(), wantq, got)
		}
		if got := r.big(); got.Cmp(wantr) != 0 {
			t.Fatalf("%d / %d expected remainder of %d, got %d",
				x.big(), y.big(), wantr, got)
		}
	}
}

func TestUint192QuoRem64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint192()
		y := randUint64()
		if y == 0 {
			y = 1
		}

		q, r := x.quoRem64(y)

		ybig := new(big.Int).SetUint64(y)
		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), ybig, wantr)
		wantq.And(wantq, big192mask)

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

func TestUint192Lsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := randUint192()
		n := uint(rand.Intn(192 + 1))

		z := x.Lsh(n)

		want := new(big.Int).Lsh(x.big(), n)
		want.And(want, big192mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d << %d: expected %d, got %d",
				x.big(), n, want, got)
		}
	}
}

func TestUint192Rsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := randUint192()
		n := uint(rand.Intn(192 + 1))

		z := x.Rsh(n)

		want := new(big.Int).Rsh(x.big(), n)
		want.And(want, big192mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d >> %d: expected %d, got %d",
				x.big(), n, want, got)
		}
	}
}

func TestUint192String(t *testing.T) {
	test := func(x Uint192) {
		want := x.big().String()
		got := x.String()
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
	test(Uint192{})  // min
	test(maxUint192) // max
	for i := 0; i < 100_000; i++ {
		test(randUint192())
	}
}

func TestParseUint192(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		want := randUint192()
		b := want.big()
		for base := 2; base <= 36; base++ {
			s := b.Text(base)
			got, err := ParseUint192(s, base)
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

func BenchmarkUint192Mul(b *testing.B) {
	s := make([]Uint192, 1000)
	for i := range s {
		s[i] = randUint192()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%len(s)]
		y := s[(i+1)%len(s)]
		sink.Uint192 = x.Mul(y)
	}
}

func BenchmarkUint192QuoRem64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink.Uint192, sink.uint64 = U192(uint64(i + 2)).quoRem64(uint64(i + 1))
	}
}
