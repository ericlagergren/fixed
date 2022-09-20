//go:build ignore

package fixed

import (
	"math/big"
	"math/bits"
	"testing"

	"golang.org/x/exp/rand"
)

var (
	big256mask    = bigMask(256)
	bigMaxUint256 = Uint256{}.max().big()
)

func randUint256() Uint256 {
	var x Uint256
	if randBool() {
		x.u0 = rand.Uint64()
	}
	if randBool() {
		x.u1 = rand.Uint64()
	}
	if randBool() {
		x.u2 = rand.Uint64()
	}
	if randBool() {
		x.u3 = rand.Uint64()
	}
	return x
}

func (x Uint256) big() *big.Int {
	var v big.Int
	if bits.UintSize == 32 {
		v.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u0 >> 32),
			big.Word(x.u1),
			big.Word(x.u1 >> 32),
			big.Word(x.u2),
			big.Word(x.u2 >> 32),
			big.Word(x.u3),
			big.Word(x.u3 >> 32),
		})
	} else {
		v.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u1),
			big.Word(x.u2),
			big.Word(x.u3),
		})
	}
	return &v
}

func TestUint256Cmp(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint256()
		y := randUint256()

		got := x.Cmp(y)
		want := x.big().Cmp(y.big())
		if got != want {
			t.Fatalf("cmp(%d, %d): expected %d, got %d",
				x.big(), y.big(), want, got)
		}
	}
}

func TestUint256Add(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint256()
		y := randUint256()

		z, c := x.AddCheck(y)

		want := new(big.Int).Add(x.big(), y.big())
		if carry := want.Cmp(bigMaxUint256) > 0; carry != (c == 1) {
			t.Fatalf("%d + %d: expected %t, got %t",
				x.big(), y.big(), carry, c == 1)
		}
		want.And(want, big256mask)

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

func TestUint256Sub(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint256()
		y := randUint256()

		z, b := x.SubCheck(y)

		want := new(big.Int).Sub(x.big(), y.big())
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%d - %d: expected %t, got %t",
				x.big(), y.big(), borrow, b == 1)
		}
		want.And(want, big256mask)

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

func TestUint256Mul(t *testing.T) {
	for i := 0; i < 750_000; i++ {
		x := randUint256()
		y := randUint256()

		z, ok := x.MulCheck(y)

		want := new(big.Int).Mul(x.big(), y.big())
		if (want.Cmp(bigMaxUint256) <= 0) != ok {
			t.Fatalf("%d: %d * %d: expected %t",
				i, x.big(), y.big(), !ok)
		}
		want.And(want, big256mask)

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

func TestUint256Mul128(t *testing.T) {
	for i := 0; i < 500_000; i++ {
		x := randUint256()
		y := randUint128()

		z := x.mul128(y)

		want := new(big.Int).Mul(x.big(), y.big())
		want.And(want, big256mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d: %d * %d: expected %d, got %d",
				i, x.big(), y.big(), want, got)
		}
	}
}

func TestUint256QuoRem(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := randUint256()
		y := randUint256()
		if y == (Uint256{}) {
			y = U256(1)
		}

		q, r := x.QuoRem(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big256mask)

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

func TestUint256Lsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := randUint256()
		n := uint(rand.Intn(256 + 1))

		z := x.Lsh(n)

		want := new(big.Int).Lsh(x.big(), n)
		want.And(want, big256mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d << %d: expected %d, got %d",
				x.big(), n, want, got)
		}
	}
}

func TestUint256Rsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := randUint256()
		n := uint(rand.Intn(256 + 1))

		z := x.Rsh(n)

		want := new(big.Int).Rsh(x.big(), n)
		want.And(want, big256mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%d >> %d: expected %d, got %d",
				x.big(), n, want, got)
		}
	}
}

func TestUint256String(t *testing.T) {
	test := func(x Uint256) {
		want := x.big().String()
		got := x.String()
		if want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
	test(Uint256{})  // min
	test(maxUint256) // max
	for i := 0; i < 100_000; i++ {
		test(randUint256())
	}
}

func TestParseUint256(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		want := randUint256()
		b := want.big()
		for base := 2; base <= 36; base++ {
			s := b.Text(base)
			got, err := ParseUint256(s, base)
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

func BenchmarkUint256Mul(b *testing.B) {
	s := make([]Uint256, 1000)
	for i := range s {
		s[i] = randUint256()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%len(s)]
		y := s[(i+1)%len(s)]
		sink.Uint256 = x.Mul(y)
	}
}

func BenchmarkUint256QuoRem64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink.Uint256, sink.uint64 = U256(uint64(i + 2)).quoRem64(uint64(i + 1))
	}
}
