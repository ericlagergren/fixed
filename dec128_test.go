package fixed

import (
	"testing"

	"golang.org/x/exp/rand"
)

func d128(sign int, x uint64, exp int) Dec128 {
	return d128x(sign, U96(x), exp)
}

func d128x(sign int, x Uint96, exp int) Dec128 {
	return Dec128{x, makeFlags(sign == 1, exp)}
}

func randDec128() Dec128 {
	coeff := randUint96()
	d := coeff.digits() - 1
	for {
		exp := rand.Intn(maxExp)
		if randBool() {
			exp = -exp
		}
		adj := exp + d
		if adj >= minExp && adj <= maxExp {
			return Dec128{coeff, makeFlags(randBool(), exp)}
		}
	}
}

func TestDec128Abs(t *testing.T) {
	for i, tc := range []struct {
		in, out string
	}{
		{"-0", "0"},
		{"2.1", "2.1"},
		{"-100", "100"},
		{"101.5", "101.5"},
		{"-101.5", "101.5"},
	} {
		d, err := ParseDec128(tc.in, 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, tc.in, err)
		}
		if got := d.Abs().String(); got != tc.out {
			t.Fatalf("#%d: %q: expected %q, got %q",
				i, tc.in, tc.out, got)
		}
	}
}

func TestDec128AddSub(t *testing.T) {
	for i, tc := range []struct {
		x, y, z string
	}{
		{"12", "7.00", "19.00"},
		{"1e+2", "1e+4", "1.01e+4"},
		{"123.456", "654.321", "777.777"},
		{"1.3", "-1.07", "0.23"},
		{"1.3", "-1.30", "0.00"},
		{"1.3", "-2.07", "-0.77"},
		{"7.9228162514264337593543950335e+56", "-7.9228162514264337593543950335e+56", "0e28"},
		{"7.9228162514264337593543950335e-56", "0", "7.9228162514264337593543950335e-56"},
	} {
		x, err := ParseDec128(tc.x, 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, tc.x, err)
		}
		y, err := ParseDec128(tc.y, 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, tc.y, err)
		}
		z, err := ParseDec128(tc.z, 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, tc.z, err)
		}
		// Compare with != because we want to make sure the
		// exponent is correct.
		sum := x.Add(y)
		if sum != z {
			t.Fatalf("#%d: %q + %q: expected %#v, got %#v",
				i, tc.x, tc.y, z, sum)
		}
		// x + y = y + x
		if y.Add(x) != sum {
			t.Fatalf("#%d: %q + %q: expected %#v, got %#v",
				i, tc.y, tc.x, sum, y.Add(x))
		}
		// Compare with Equal because == won't be commutative (or
		// whatver the word I'm looking for is).
		if d := sum.Sub(x); !d.Equal(y) {
			t.Fatalf("#%d: %q - %q: expected %#v, got %#v",
				i, sum, x, y, d)
		}
		if d := sum.Sub(y); !d.Equal(x) {
			t.Fatalf("#%d: %q + %q: expected %#v, got %#v",
				i, sum, y, x, d)
		}
		println()
	}
}

func TestDec128Flags(t *testing.T) {
	for _, sign := range []bool{true, false} {
		for exp := minExp; exp < maxExp; exp++ {
			f := makeFlags(sign, exp)
			if f.signbit() != sign {
				t.Fatalf("expected %t", sign)
			}
			if f.exp() != exp {
				t.Fatalf("expected %d, got %d", exp, f.exp())
			}
		}
	}
}

func TestParseDec128(t *testing.T) {
	for i, tc := range []struct {
		in  string
		out Dec128
	}{
		{"0", d128(0, 0, 0)},
		{"0.00", d128(0, 0, -2)},
		{"123", d128(0, 123, 0)},
		{"-123", d128(1, 123, 0)},
		{"1.23E3", d128(0, 123, 1)},
		{"1.23E+3", d128(0, 123, 1)},
		{"12.3E+7", d128(0, 123, 6)},
		{"12.0", d128(0, 120, -1)},
		{"12.3", d128(0, 123, -1)},
		{"0.00123", d128(0, 123, -5)},
		{"-1.23E-12", d128(1, 123, -14)},
		{"1234.5E-4", d128(0, 12345, -5)},
		{"-0", d128(1, 0, 0)},
		{"-0.00", d128(1, 0, -2)},
		{"0E+7", d128(0, 0, 7)},
		{"-0E-7", d128(1, 0, -7)},
		{"7.9228162514264337593543950335e+56", d128x(0, Uint96{}.max(), 28)},
		{"-7.9228162514264337593543950335e+56", d128x(1, Uint96{}.max(), 28)},
		{"7.9228162514264337593543950335e-56", d128x(0, Uint96{}.max(), -84)},
		{"-7.9228162514264337593543950335e-56", d128x(1, Uint96{}.max(), -84)},
	} {
		got, err := ParseDec128(tc.in, 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, tc.in, err)
		}
		if tc.out != got {
			t.Fatalf("#%d: %q: expected %#v, got %#v",
				i, tc.in, tc.out, got)
		}
	}

	for i := 0; i < 100_000; i++ {
		want := randDec128()
		got, err := ParseDec128(want.String(), 10)
		if err != nil {
			t.Fatalf("#%d: %q: %v", i, want, err)
		}
		if want != got {
			t.Fatalf("#%d: %q: expected %#v, got %#v",
				i, want, want, got)
		}
	}
}

var dec128StrTests = []struct {
	d Dec128
	s string
}{
	{d128(0, 123, 0), "123"},
	{d128(1, 123, 0), "-123"},
	{d128(0, 123, 1), "1.23e+3"},
	{d128(0, 123, 3), "1.23e+5"},
	{d128(0, 123, -1), "12.3"},
	{d128(0, 123, -5), "0.00123"},
	{d128(0, 123, -10), "1.23e-8"},
	{d128(1, 123, -12), "-1.23e-10"},
	{d128(0, 0, 0), "0"},
	{d128(0, 0, -2), "0.00"},
	{d128(0, 0, 2), "0e+2"},
	{d128(1, 0, 0), "-0"},
	{d128(0, 5, -6), "0.000005"},
	{d128(0, 50, -7), "0.0000050"},
	{d128(0, 5, -7), "5e-7"},
}

func TestDec128String(t *testing.T) {
	for i, tc := range dec128StrTests {
		got := tc.d.String()
		if got != tc.s {
			t.Fatalf("#%d: %#v: expected %q, got %q",
				i, tc.d, tc.s, got)
		}
	}
}

// TestDec128StringAllocs checks that String does not allocate
// more than expected.
func TestDec128StringAllocs(t *testing.T) {
	for i, tc := range dec128StrTests {
		n := int(testing.AllocsPerRun(10, func() {
			sink.string = tc.d.String()
		}))
		if n > 1 {
			t.Fatalf("#%d: %#v: expected <= %d, got %d",
				i, tc.d, 1, n)
		}
	}
}

func BenchmarkDec128Add(b *testing.B) {
	s := make([]Dec128, 100)
	for i := range s {
		s[i] = randDec128()
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink.dec128 = s[i%len(s)].Add(s[(i+1)%len(s)])
	}
}

func BenchmarkDec128String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := dec128StrTests[i%len(dec128StrTests)]
		sink.string = tc.d.String()
	}
}
