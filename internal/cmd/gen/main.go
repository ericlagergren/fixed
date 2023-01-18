package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

func main() {
	for _, s := range os.Args[1:] {
		if err := main1(s); err != nil {
			log.Fatal(err)
		}
	}
}

func main1(s string) error {
	bits, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if bits < 0 || bits%64 != 0 {
		return fmt.Errorf("invalid bit size: %d", bits)
	}
	for _, v := range []struct {
		path string
		fn   func(*bytes.Buffer, int)
	}{
		{fmt.Sprintf("uint%d.go", bits), gen},
		{fmt.Sprintf("uint%d_test.go", bits), genTest},
	} {
		var b bytes.Buffer
		v.fn(&b, bits)
		src, err := format.Source(b.Bytes())
		if err != nil {
			// Use original source to make it easier to debug.
			src = b.Bytes()
		}
		writeErr := os.WriteFile(v.path, src, 0644)
		// Prefer write errors to code formatting errors.
		if writeErr != nil {
			return writeErr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type namedArg struct {
	name  string
	value any
}

func named(name string, value any) namedArg {
	return namedArg{name, value}
}

func fprintf(w io.Writer, format string, args ...any) {
	named := make(map[string]any)
	var newArgs []any
	for _, a := range args {
		if v, ok := a.(namedArg); ok {
			if _, ok := named[v.name]; ok {
				panic("duplicate arg: " + v.name)
			}
			named[v.name] = v.value
			continue
		}
		if len(named) != 0 {
			panic("positional arg after named arg")
		}
		newArgs = append(newArgs, a)
	}
	var b strings.Builder
	for s := format; s != ""; {
		i := strings.Index(s, "{:")
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])

		j := strings.IndexByte(s[i+2:], '}')
		if j < 0 {
			panic("unclosed curly brace")
		}
		name := s[i+2 : i+2+j]
		if _, ok := named[name]; !ok {
			panic("unknown name: " + name)
		}
		fmt.Fprintf(&b, "%v", named[name])
		s = s[i+2+j+1:]
	}
	fmt.Fprintf(w, b.String(), newArgs...)
}

func gen(b *bytes.Buffer, bits int) {
	p := func(format string, args ...any) {
		args = append(args,
			named("name", fmt.Sprintf("Uint%d", bits)),
			named("bits", bits),
			named("halfBits", bits/2),
			named("halfMask", (bits/2)-1),
		)
		fprintf(b, format, args...)
	}

	nelems := (bits / 64) - 1

	max := big.NewInt(1)
	max.Lsh(max, uint(bits))
	max.Sub(max, big.NewInt(1))
	v := big.NewInt(1)
	tabLen := 0
	for v.Cmp(max) <= 0 {
		v.Mul(v, big.NewInt(10))
		tabLen++
	}

	p(`
// Code generated by 'gen'. DO NOT EDIT.

package fixed

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"sync"
)

// {:name} is an unsigned, {:bits}-bit integer.
//
// It can be compared for equality with ==.
type {:name} struct {
`)
	for i := 0; i < bits/64; i++ {
		if i > 0 {
			p(", ")
		}
		p("u%d", i)
	}
	p(` uint64
}

var _ Uint[{:name}] = {:name}{}

// U{:bits} returns x as a {:name}.
func U{:bits}(x uint64) {:name} {
	return {:name}{u0: x}
}

func u{:bits}(lo, hi Uint{:halfBits}) {:name} {
	return {:name}{`)
	for _, s := range []string{"lo", "hi"} {
		n := (bits / 64) / 2
		for i := 0; i < n; i++ {
			if n >= 8 {
				p("\n")
			}
			p("%s.u%d,", s, i)
		}
	}
	p(`
	}
}

func ({:name}) max() {:name} {
	return {:name}{
`)
	for i := 0; i < bits/64; i++ {
		p("math.MaxUint64,\n")
	}
	p(`}
}

func (x {:name}) low() Uint%[1]d {
	return Uint%[1]d {`, bits/2)
	n := (bits / 64) / 2
	for i := 0; i < n; i++ {
		if n >= 8 {
			p("\n")
		}
		p("x.u%d,", i)
	}
	p(`}
}

func (x {:name}) high() Uint%[1]d {
	return Uint%[1]d {`, bits/2)
	n = (bits / 64) / 2
	for i := n; i < bits/64; i++ {
		if n >= 8 {
			p("\n")
		}
		p("x.u%d,", i)
	}
	p(`}
}

func (x {:name}) uint8() uint8 {
	return uint8(x.u0)
}

// Bytes returns x encoded as a little-endian integer.
func (x {:name}) Bytes() []byte {
`)
	p("b := make([]byte, %d)\n", (bits+7)/8)
	for i := 0; i < bits/64; i++ {
		p("binary.LittleEndian.PutUint64(b[%d:], x.u%d)\n",
			i*8, i)
	}
	p(`return b
}

// SetBytes sets x to the encoded little-endian integer b.
func (x *{:name}) SetBytes(b []byte) error {
`)
	p("if len(b) != %d {", (bits+7)/8)
	p(`return fmt.Errorf("fixed: invalid length: %%d", len(b))
	}
`)
	for i := 0; i < bits/64; i++ {
		p("x.u%d = binary.LittleEndian.Uint64(b[%d:])\n",
			i, i*8)
	}
	p(`return nil
}

// Size returns the width of the integer in bits.
func ({:name}) Size() int {
	return {:bits}
}

// BitLen returns the number of bits required to represent x.
func (x {:name}) BitLen() int {
	switch {
`)
	for i := nelems; i > 0; i-- {
		p("case x.u%d != 0:\n", i)
		p("return %d + bits.Len64(x.u%d)\n", i*64, i)
	}
	p(`default: return bits.Len64(x.u0) }
}

// LeadingZeros returns the number of leading zeros in x.
func (x {:name}) LeadingZeros() int {
	return {:bits} - x.BitLen()
}

// IsZero is shorthand for x == {:name}{}.
func (x {:name}) IsZero() bool {
	return x == {:name}{}
}

// Cmp compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x {:name}) Cmp(y {:name}) int {
	switch {
`)
	for i := nelems; i > 0; i-- {
		p("case x.u%d != y.u%d:\n", i, i)
		p("return cmp(x.u%d, y.u%d)\n", i, i)
	}
	p(`default:
		return cmp(x.u0, y.u0)
	}
}

// cmp64 compares x and y and returns
//
//   - +1 if x > y
//   - 0 if x == y
//   - -1 if x < y
func (x {:name}) cmp64(y uint64) int {
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
func (x {:name}) Equal(y {:name}) bool {
	return x == y
}

// And returns x&y.
func (x {:name}) And(y {:name}) {:name} {
	return {:name}{
`)
	for i := 0; i < bits/64; i++ {
		p("x.u%d & y.u%d,\n", i, i)
	}
	p(`
	}
}

// Or returns x|y.
func (x {:name}) Or(y {:name}) {:name} {
	return {:name}{
`)
	for i := 0; i < bits/64; i++ {
		p("x.u%d | y.u%d,\n", i, i)
	}
	p(`
	}
}

// orLsh64 returns x | y<<s.
func (x {:name}) orLsh64(y uint64, s uint) {:name} {
	return x.Or({:name}{u0: y}.Lsh(s))
}

// Xor returns x^y.
func (x {:name}) Xor(y {:name}) {:name} {
	return {:name}{
`)
	for i := 0; i < bits/64; i++ {
		p("x.u%d ^ y.u%d,\n", i, i)
	}
	p(`
	}
}

// Lsh returns x<<n.
func (x {:name}) Lsh(n uint) {:name} {
	switch {`)
	for i := nelems; i > 0; i-- {
		p(`
	case n > %[1]d:
		s := n - %[1]d
`, i*64)
		if i < nelems {
			p("ŝ := 64 - s\n")
		}
		p("return {:name}{\n")
		p("u%d: x.u0 << s,\n", i)
		for j := i + 1; j < bits/64; j++ {
			p("u%d: x.u%d<<s | x.u%d>>ŝ,\n", j, j-i, j-i-1)
		}
		p("}")
	}
	p(`
	default:
		s := n
		ŝ := 64 - s
		return {:name}{
`)
	p("u0: x.u0<<s,\n")
	for i := 1; i < bits/64; i++ {
		p("u%d: x.u%d<<s | x.u%d>>ŝ,\n", i, i, i-1)
	}
	p(`
		}
	}
}

// Rsh returns x>>n.
func (x {:name}) Rsh(n uint) {:name} {
	switch {`)
	for i := nelems; i > 0; i-- {
		p(`
	case n > %[1]d:
		s := n - %[1]d
`, i*64)
		if i < nelems {
			p("ŝ := 64 - s\n")
		}
		p("return {:name}{\n")
		j, k := i, 0
		for k < nelems-i {
			p("u%d: x.u%d>>s | x.u%d<<ŝ,\n", k, j, j+1)
			k++
			j++
		}
		p("u%d: x.u%d >> s,\n", k, j)
		p("}")
	}
	p(`
	default:
		s := n
		ŝ := 64 - s
		return {:name}{
`)
	for j := 0; j < nelems; j++ {
		p("u%[1]d: x.u%[1]d>>s | x.u%[2]d<<ŝ,\n", j, j+1)
	}
	p("u%[1]d: x.u%[1]d >> s,\n", nelems)
	p(`
		}
	}
}

// Add returns x+y.
func (x {:name}) Add(y {:name}) {:name} {
	var z {:name}
	var carry uint64
`)
	p("z.u0, carry = bits.Add64(x.u0, y.u0, 0)\n")
	for i := 1; i < nelems; i++ {
		p("z.u%[1]d, carry = bits.Add64(x.u%[1]d, y.u%[1]d, carry)\n", i)
	}
	p("z.u%[1]d, _ = bits.Add64(x.u%[1]d, y.u%[1]d, carry)\n", nelems)
	p(`return z
}

// add64 returns x+y.
func (x {:name}) add64(y uint64) {:name} {
	var z {:name}
	var carry uint64
`)
	p("z.u0, carry = bits.Add64(x.u0, y, 0)\n")
	for i := 1; i < nelems; i++ {
		p("z.u%[1]d, carry = bits.Add64(x.u%[1]d, 0, carry)\n", i)
	}
	p("z.u%[1]d, _ = bits.Add64(x.u%[1]d, 0, carry)\n", nelems)
	p(`return z
}

// AddCheck returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x {:name}) AddCheck(y {:name}) (z {:name}, carry uint64) {
`)
	p("z.u0, carry = bits.Add64(x.u0, y.u0, 0)\n")
	for i := 1; i < bits/64; i++ {
		p("z.u%[1]d, carry = bits.Add64(x.u%[1]d, y.u%[1]d, carry)\n", i)
	}
	p(`return z, carry
}

// addCheck64 returns x+y.
//
// carry is 1 if x+y overflows and 0 otherwise.
func (x {:name}) addCheck64(y uint64) (z {:name}, carry uint64) {
`)
	p("z.u0, carry = bits.Add64(x.u0, y, 0)\n")
	for i := 1; i < bits/64; i++ {
		p("z.u%[1]d, carry = bits.Add64(x.u%[1]d, 0, carry)\n", i)
	}
	p(`return z, carry
}

// Sub returns x-y.
func (x {:name}) Sub(y {:name}) {:name} {
	var z {:name}
	var borrow uint64
`)
	p("z.u0, borrow = bits.Sub64(x.u0, y.u0, 0)\n")
	for i := 1; i < nelems; i++ {
		p("z.u%[1]d, borrow = bits.Sub64(x.u%[1]d, y.u%[1]d, borrow)\n", i)
	}
	p("z.u%[1]d, _ = bits.Sub64(x.u%[1]d, y.u%[1]d, borrow)\n", nelems)
	p(`return z
}

// sub64 returns x-y.
func (x {:name}) sub64(y uint64) {:name} {
	var z {:name}
	var borrow uint64
`)
	p("z.u0, borrow = bits.Sub64(x.u0, y, 0)\n")
	for i := 1; i < nelems; i++ {
		p("z.u%[1]d, borrow = bits.Sub64(x.u%[1]d, 0, borrow)\n", i)
	}
	p("z.u%[1]d, _ = bits.Sub64(x.u%[1]d, 0, borrow)\n", nelems)
	p(`return z
}

// SubCheck returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x {:name}) SubCheck(y {:name}) (z {:name}, borrow uint64) {
`)
	p("z.u0, borrow = bits.Sub64(x.u0, y.u0, 0)\n")
	for i := 1; i < bits/64; i++ {
		p("z.u%[1]d, borrow = bits.Sub64(x.u%[1]d, y.u%[1]d, borrow)\n", i)
	}
	p(`return z, borrow
}

// subCheck64 returns x-y.
//
// borrow is 1 if x-y overflows and 0 otherwise.
func (x {:name}) subCheck64(y uint64) (z {:name}, borrow uint64) {
`)
	p("z.u0, borrow = bits.Sub64(x.u0, y, 0)\n")
	for i := 1; i < bits/64; i++ {
		p("z.u%[1]d, borrow = bits.Sub64(x.u%[1]d, 0, borrow)\n", i)
	}
	p(`return z, borrow
}

// Mul returns x*y.
func (x {:name}) Mul(y {:name}) {:name} {
	var z {:name}
	var c uint64

`)
	for yi := 0; yi < bits/64; yi++ {
		p("// y.u%d * x\n", yi)
		p("if d := y.u%d; d != 0 {\n", yi)

		xi := 0
		if yi == 0 {
			p("c, z.u0 = bits.Mul64(x.u0, d)\n")
			xi++
		} else if yi < nelems {
			p("c, z.u%[1]d = mulAddWWW(x.u0, d, z.u%[1]d)\n", yi)
			xi++
		}

		for zi := yi + 1; zi < nelems; zi++ {
			if yi == 0 {
				p("c, z.u%[1]d = mulAddWWW(x.u%[2]d, d, c)\n", zi, xi)
			} else {
				p("c, z.u%[1]d = mulAddWWWW(x.u%[2]d, d, z.u%[1]d, c)\n", zi, xi)
			}
			xi++
		}

		if yi == nelems {
			p("z.u%d += x.u%d*d\n", nelems, xi)
		} else {
			p("z.u%d += x.u%d*d + c\n", nelems, xi)
		}
		p("}\n\n")
	}
	p(`return z
}

func (x {:name}) mul{:halfBits}(y Uint{:halfBits}) {:name} {
	var z {:name}
	var c uint64

`)
	for yi := 0; yi < (bits/64)/2; yi++ {
		p("// y.u%d * x\n", yi)
		p("if d := y.u%d; d != 0 {\n", yi)

		xi := 0
		if yi == 0 {
			p("c, z.u0 = bits.Mul64(x.u0, d)\n")
			xi++
		} else if yi < nelems {
			p("c, z.u%[1]d = mulAddWWW(x.u0, d, z.u%[1]d)\n", yi)
			xi++
		}

		for zi := yi + 1; zi < nelems; zi++ {
			if yi == 0 {
				p("c, z.u%[1]d = mulAddWWW(x.u%[2]d, d, c)\n", zi, xi)
			} else {
				p("c, z.u%[1]d = mulAddWWWW(x.u%[2]d, d, z.u%[1]d, c)\n", zi, xi)
			}
			xi++
		}

		if yi == nelems {
			p("z.u%d += x.u%d*d\n", nelems, xi)
		} else {
			p("z.u%d += x.u%d*d + c\n", nelems, xi)
		}
		p("}\n\n")
	}
	p(`return z
}

func (x {:name}) mul64(y uint64) {:name} {
	if y == 0 {
		return {:name}{}
	}
	var z {:name}
	var c uint64
`)
	p("c, z.u0 = bits.Mul64(x.u0, y)\n")
	for xi := 1; xi < nelems; xi++ {
		p("c, z.u%[1]d = mulAddWWW(x.u%[1]d, y, c)\n", xi)
	}
	p("z.u%[1]d += x.u%[1]d*y + c\n", nelems)
	p(`return z
}

// Exp return x^y mod m.
//
// If m == 0, Exp simply returns x^y.
func (x {:name}) Exp(y, m {:name}) {:name} {
	const mask = 1 << (64-1)

	// x^0 = 1
	if y.IsZero() {
		return U{:bits}(1)
	}

	// x^1 mod m == x mod m
	mod := !m.IsZero()
	if y == U{:bits}(1) && mod {
		_, r := x.QuoRem(m)
		return r
	}

	yv := []uint64{
`)
	for i := 0; i < nelems; i++ {
		p("y.u%d,", i)
	}
	p(`
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
func (x {:name}) mulPow10(n uint) ({:name}, bool) {
	switch {
	case x.IsZero():
		return {:name}{}, true
	case n == 0:
		return x, true`)
	if bits <= 256 {
		p(`
	case n >= %d:
		return {:name}{}, false
	default:
		return x.MulCheck(pow10{:name}(n))`, tabLen)
	} else {
		p(`
	default:
		return x.MulCheck(U{:bits}(10).Exp(U{:bits}(uint64(n)), U{:bits}(0)))`)
	}
	p(`
	}
}

var pow10tab{:name} struct{
	values []{:name}
	once sync.Once
}

func pow10{:name}(n uint) {:name} {
	pow10tab{:name}.once.Do(func() {
		tab := make([]{:name}, 2 + %d)
		tab[0] = {:name}{}
		tab[1] = U{:bits}(1)
		for i := 2; i < len(tab); i++ {
			tab[i] = tab[i-1].mul64(10)
		}
		pow10tab{:name}.values = tab
	})
	return pow10tab{:name}.values[n]
}`, tabLen)
	p(`

// MulCheck returns x*y and reports whether the multiplication
// oveflowed.
func (x {:name}) MulCheck(y {:name}) ({:name}, bool) {
	if x.BitLen()+y.BitLen() > {:bits} {
		return {:name}{}, false
	}

	var z {:name}
	var c uint64

`)
	for yi := 0; yi < bits/64; yi++ {
		p("// y.u%d * x\n", yi)
		p("if d := y.u%d; d != 0 {\n", yi)

		if yi == 0 {
			p("c, z.u0 = bits.Mul64(x.u0, d)\n")
		} else {
			p("c, z.u%[1]d = mulAddWWW(x.u0, d, z.u%[1]d)\n", yi)
		}
		xi := 1

		for zi := yi + 1; zi < bits/64; zi++ {
			if yi == 0 {
				p("c, z.u%[1]d = mulAddWWW(x.u%[2]d, d, c)\n", zi, xi)
			} else {
				p("c, z.u%[1]d = mulAddWWWW(x.u%[2]d, d, z.u%[1]d, c)\n", zi, xi)
			}
			xi++
		}
		p("if c != 0 {\n")
		p("return {:name}{}, false\n")
		p("}\n")

		p("}\n\n")
	}
	p(`return z, true
}

func (x {:name}) mulCheck64(y uint64) ({:name}, bool) {
	if y == 0 {
		return {:name}{}, true
	}
	var z {:name}
	var c uint64
`)
	p("c, z.u0 = bits.Mul64(x.u0, y)\n")
	for xi := 1; xi < bits/64; xi++ {
		p("c, z.u%[1]d = mulAddWWW(x.u%[1]d, y, c)\n", xi)
	}
	p(`if c != 0 {
		return {:name}{}, false
	}
	return z, true
}

// QuoRem returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x {:name}) QuoRem(y {:name}) (q, r {:name}) {
	if x.Cmp(y) < 0 {
		// x/y for x < y = 0.
		// x%%y for x < y = x.
		return {:name}{}, x
	}

	if y.high().IsZero() {
		q, rr := x.quoRem{:halfBits}(y.low())
		return q, u{:bits}(rr, Uint{:halfBits}{})
	}

	n := uint(y.high().LeadingZeros())
	y1 := y.Lsh(n) // y1 := y<<n
	x1 := x.Rsh(1) // x1 := x>>1
	tq, _ := div{:halfBits}(x1.high(), x1.low(), y1.high())
	tq = tq.Rsh({:halfMask} - n) // tq >>= {:halfMask} - n
	if !tq.IsZero() {
		tq = tq.sub64(1) // tq--
	}
	q = u{:bits}(tq, Uint{:halfBits}{})
	ytq := y.mul{:halfBits}(tq) // ytq := y*tq
	r = x.Sub(ytq)      // r = x-ytq
	if r.Cmp(y) >= 0 {
		q = q.add64(1) // q++
		r = r.Sub(y)   // r -= y
	}
	return
}

// quoRem{:halfBits} returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x {:name}) quoRem{:halfBits}(y Uint{:halfBits}) (q {:name}, r Uint{:halfBits}) {
	if x.high().Cmp(y) < 0 {
		lo, r := div{:halfBits}(x.high(), x.low(), y)
		return u{:bits}(lo, Uint{:halfBits}{}), r
	}
	hi, r := div{:halfBits}(Uint{:halfBits}{}, x.high(), y)
	lo, r := div{:halfBits}(r, x.low(), y)
	return u{:bits}(lo, hi), r
}

// quoRem64 returns (q, r) such that
//
//	q = x/y
//	r = x - y*q
func (x {:name}) quoRem64(y uint64) (q {:name}, r uint64) {
`)
	p("q.u%[1]d, r = bits.Div64(0, x.u%[1]d, y)\n", nelems)
	for i := (bits / 64) - 2; i >= 0; i-- {
		p("q.u%[1]d, r = bits.Div64(r, x.u%[1]d, y)\n", i)
	}
	p(`return q, r
}

// div{:bits} returns (q, r) such that
//
//  q = (hi, lo)/y
//  r = (hi, lo) - y*q
func div{:bits}(hi, lo, y Uint{:bits}) (q, r Uint{:bits}) {
	if y.IsZero() {
		panic("integer divide by zero")
	}
	if y.Cmp(hi) <= 0 {
		panic("integer overflow")
	}

	s := uint(y.LeadingZeros())
	y = y.Lsh(s) // y = y<<s
	yn1 := y.high() // yn1 := y >> {:halfBits}
	yn0 := y.low() // yn0 := y & mask{:halfBits}

	un32 := hi.Lsh(s).Or(lo.Rsh({:bits} - s)) // un32 := hi<<s | lo>>({:bits}-s)
	un10 := lo.Lsh(s) // un10 := lo<<s
	un1 := un10.high() // un1 := un10 >> {:halfBits}
	un0 := un10.low() // un0 := un10 & mask{:halfBits}
	q1, rhat := un32.quoRem{:halfBits}(yn1)

	var c uint64 // rhat + yn1 carry

	// for q1 >= two{:halfBits} || q1*yn0 > two{:halfBits}*rhat+un1 { ... }
	for !q1.high().IsZero() || q1.mul{:halfBits}(yn0).Cmp(u{:bits}(un1, rhat)) > 0 {
		q1 = q1.sub64(1) // q1--
		rhat, c = rhat.AddCheck(yn1) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// un21 := un32*two{:halfBits} + un1 - q1*y
	un21 := u{:bits}(un1, un32.low()).Sub(q1.Mul(y))
	q0, rhat := un21.quoRem{:halfBits}(yn1)

	// for q0 >= two{:halfBits} || q0*yn0 > two{:halfBits}*rhat+un0 { ... }
	for !q0.high().IsZero() || q0.mul{:halfBits}(yn0).Cmp(u{:bits}(un0, rhat)) > 0 {
		q0 = q0.sub64(1) // q0--
		rhat, c = rhat.AddCheck(yn1) // rhat += yn1
		if c != 0 {
			break
		}
	}

	// q = q1*two{:halfBits} + q0
	q = u{:bits}(q0.low(), q1.low())
	// r = (un21*two{:halfBits} + un0 - q0*y) >> s
	r = u{:bits}(un0, un21.low()).Sub(q0.Mul(y)).Rsh(s)
	return
}

func (x {:name}) GoString() string {
	return fmt.Sprintf("[`)
	for i := 0; i < bits/64; i++ {
		if i > 0 {
			p(" ")
		}
		p("%%d")
	}
	p("]\",\n")
	for i := 0; i < bits/64; i++ {
		p("x.u%d,\n", i)
	}
	p(`)
}

// String returns the base-10 representation of x.
func (x {:name}) String() string {
`)
	p("b := make([]byte, %d)\n", maxStrLen(uint(bits)))
	p(`i := len(b)
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

// Parse{:name} returns the value of s in the given base.
func Parse{:name}(s string, base int) ({:name}, error) {
	x, _, _, err := parseUint[{:name}](s, base, false)
	return x, err
}
`)
}

func maxStrLen(bits uint) int {
	v := new(big.Int).Lsh(big.NewInt(1), bits) // 1<<n
	v.Sub(v, big.NewInt(1))                    // 1<<n - 1
	return len(v.String())
}

func genTest(b *bytes.Buffer, bits int) {
	p := func(format string, args ...any) {
		args = append(args,
			named("name", fmt.Sprintf("Uint%d", bits)),
			named("bits", bits),
			named("halfBits", bits/2),
		)
		fprintf(b, format, args...)
	}

	p(`// Code generated by 'gen'. DO NOT EDIT.

package fixed

import (
	"math/big"
	"math/bits"
	"testing"

	"golang.org/x/exp/rand"
)

var (
	big{:bits}mask = bigMask({:bits})
	bigMax{:name} = {:name}{}.max().big()
)

func rand{:name}() {:name} {
	var x {:name}
`)
	for i := 0; i < bits/64; i++ {
		p("if randBool() { x.u%d = randUint64() }\n", i)
	}
	p(`return x
}

func (x {:name}) big() *big.Int {
	var v big.Int
	if bits.UintSize == 32 {
		v.SetBits([]big.Word{
`)
	for i := 0; i < bits/64; i++ {
		p("big.Word(x.u%d),\n", i)
		p("big.Word(x.u%d >> 32),\n", i)
	}
	p(`
		})
	} else {
		v.SetBits([]big.Word{
`)
	for i := 0; i < bits/64; i++ {
		p("big.Word(x.u%d),\n", i)
	}
	p(`
		})
	}
	return &v
}

func Test{:name}BitLen(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := rand{:name}()

		got := x.BitLen()
		want := x.big().BitLen()
		if got != want {
			t.Fatalf("expected %%d, got %%d", want, got)
		}
	}
}

func Test{:name}LeadingZeros(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := rand{:name}()

		got := x.LeadingZeros()
		want := {:bits} - x.big().BitLen()
		if got != want {
			t.Fatalf("expected %%d, got %%d", want, got)
		}
	}
}

func Test{:name}Cmp(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		got := x.Cmp(y)
		want := x.big().Cmp(y.big())
		if got != want {
			t.Fatalf("Cmp(%%d, %%d): expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}And(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z := x.And(y)

		want := new(big.Int).And(x.big(), y.big())
		want.And(want, big{:bits}mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d + %%d: expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Or(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z := x.Or(y)

		want := new(big.Int).Or(x.big(), y.big())
		want.And(want, big{:bits}mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d + %%d: expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Xor(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z := x.Xor(y)

		want := new(big.Int).Xor(x.big(), y.big())
		want.And(want, big{:bits}mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d + %%d: expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Lsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := rand{:name}()
		n := uint(rand.Intn({:bits} + 1))

		z := x.Lsh(n)

		want := new(big.Int).Lsh(x.big(), n)
		want.And(want, big{:bits}mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d << %%d: expected %%d, got %%d",
				x.big(), n, want, got)
		}
	}
}

func Test{:name}Rsh(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		x := rand{:name}()
		n := uint(rand.Intn({:bits} + 1))

		z := x.Rsh(n)

		want := new(big.Int).Rsh(x.big(), n)
		want.And(want, big{:bits}mask)

		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d >> %%d: expected %%d, got %%d",
				x.big(), n, want, got)
		}
	}
}

func Test{:name}Add(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z, c := x.AddCheck(y)

		want := new(big.Int).Add(x.big(), y.big())
		if carry := want.Cmp(bigMax{:name}) > 0; carry != (c == 1) {
			t.Fatalf("%%d + %%d: expected %%t, got %%t",
				x.big(), y.big(), carry, c == 1)
		}
		want.And(want, big{:bits}mask)

		if c == 0 && x.Add(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), y.big(), x.Add(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d + %%d: expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Add64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := rand{:name}()
		y := randUint64()

		z, c := x.addCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Add(x.big(), ybig)
		if carry := want.Cmp(bigMax{:name}) > 0; carry != (c == 1) {
			t.Fatalf("%%d + %%d: expected %%t, got %%t",
				x.big(), ybig, carry, c == 1)
		}
		want.And(want, big{:bits}mask)

		if c == 0 && x.add64(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), ybig, x.add64(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d + %%d: expected %%d, got %%d",
				x.big(), ybig, want, got)
		}
	}
}

func Test{:name}Sub(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z, b := x.SubCheck(y)

		want := new(big.Int).Sub(x.big(), y.big())
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%%d - %%d: expected %%t, got %%t",
				x.big(), y.big(), borrow, b == 1)
		}
		want.And(want, big{:bits}mask)

		if b == 0 && x.Sub(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), y.big(), x.Sub(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d - %%d: expected %%d, got %%d",
				x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Sub64(t *testing.T) {
	for i := 0; i < 250_000; i++ {
		x := rand{:name}()
		y := randUint64()

		z, b := x.subCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Sub(x.big(), ybig)
		if borrow := want.Sign() < 0; borrow != (b == 1) {
			t.Fatalf("%%d - %%d: expected %%t, got %%t",
				x.big(), ybig, borrow, b == 1)
		}
		want.And(want, big{:bits}mask)

		if b == 0 && x.sub64(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), ybig, x.sub64(y), z)
		}
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d - %%d: expected %%d, got %%d",
				x.big(), ybig, want, got)
		}
	}
}

func Test{:name}Mul(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()

		z, ok := x.MulCheck(y)

		want := new(big.Int).Mul(x.big(), y.big())
		if (want.Cmp(bigMax{:name}) <= 0) != ok {
			t.Fatalf("%%d: %%d * %%d: expected %%t",
				i, x.big(), y.big(), !ok)
		}
		want.And(want, big{:bits}mask)

		if ok && x.Mul(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), y.big(), x.Mul(y), z)
		}
		z = x.Mul(y)
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d: %%d * %%d: expected %%d, got %%d",
				i, x.big(), y.big(), want, got)
		}
	}
}

func Test{:name}Mul64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := randUint64()

		z, ok := x.mulCheck64(y)

		ybig := new(big.Int).SetUint64(y)
		want := new(big.Int).Mul(x.big(), ybig)
		if (want.Cmp(bigMax{:name}) <= 0) != ok {
			t.Fatalf("%%d: %%d * %%d: expected %%t",
				i, x.big(), ybig, !ok)
		}
		want.And(want, big{:bits}mask)

		if ok && x.mul64(y) != z {
			t.Fatalf("%%d: %%d * %%d: %%d != %%d",
				i, x.big(), ybig, x.mul64(y), z)
		}
		z = x.mul64(y)
		if got := z.big(); got.Cmp(want) != 0 {
			t.Fatalf("%%d: %%d * %%d: expected %%d, got %%d",
				i, x.big(), ybig, want, got)
		}
	}
}

func Test{:name}Exp(t* testing.T) {
	x := U{:bits}(1)
	ten := U{:bits}(10)
	for i := 1;; i++ {
		want, ok := x.mulCheck64(10)
		if !ok { break }
		got := ten.Exp(U{:bits}(uint64(i)), U{:bits}(0))
		if got != want {
			t.Fatalf("#%%d: expected %%q, got %%q", i, want, got)
		}
		x = want
	}
}

func Test{:name}QuoRem(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := rand{:name}()
		if y.IsZero() {
			y = U{:bits}(1)
		}

		q, r := x.QuoRem(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big{:bits}mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%%d / %%d expected quotient of %%d, got %%d",
				x.big(), y.big(), wantq, got)
		}
		if got := r.big(); got.Cmp(wantr) != 0 {
			t.Fatalf("%%d / %%d expected remainder of %%d, got %%d",
				x.big(), y.big(), wantr, got)
		}
	}
}

func Test{:name}QuoRemHalf(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := randUint{:halfBits}()
		if y.IsZero() {
			y = U{:halfBits}(1)
		}

		q, r := x.quoRem{:halfBits}(y)

		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), y.big(), wantr)
		wantq.And(wantq, big{:bits}mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%%d / %%d expected quotient of %%d, got %%d",
				x.big(), y.big(), wantq, got)
		}
		if got := r.big(); got.Cmp(wantr) != 0 {
			t.Fatalf("%%d / %%d expected remainder of %%d, got %%d",
				x.big(), y.big(), wantr, got)
		}
	}
}

func Test{:name}QuoRem64(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x := rand{:name}()
		y := randUint64()
		if y == 0 {
			y = 1
		}

		q, r := x.quoRem64(y)

		ybig := new(big.Int).SetUint64(y)
		wantq := new(big.Int)
		wantr := new(big.Int)
		wantq.QuoRem(x.big(), ybig, wantr)
		wantq.And(wantq, big{:bits}mask)

		if got := q.big(); got.Cmp(wantq) != 0 {
			t.Fatalf("%%d / %%d expected quotient of %%d, got %%d",
				x.big(), ybig, wantq, got)
		}
		if got := new(big.Int).SetUint64(r); got.Cmp(wantr) != 0 {
			t.Fatalf("%%d / %%d expected remainder of %%d, got %%d",
				x.big(), ybig, wantr, got)
		}
	}
}

func Test{:name}String(t *testing.T) {
	test := func(x {:name}) {
		want := x.big().String()
		got := x.String()
		if want != got {
			t.Fatalf("expected %%q, got %%q", want, got)
		}
	}
	test({:name}{}) // min
	test({:name}{}.max()) // max
	for i := 0; i < 10_000; i++ {
		test(rand{:name}())
	}
}

func TestParse{:name}(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		want := rand{:name}()
		b := want.big()
		for base := 2; base <= 36; base++ {
			s := b.Text(base)
			got, err := Parse{:name}(s, base)
			if err != nil {
				t.Fatalf("%%q in base %%d: unexpected error: %%v", s, base, err)
			}
			if got != want {
				t.Fatalf("%%q in base %%d: expected %%#v, got %%#v",
					s, base, want, got)
			}
		}
	}
}

func Benchmark{:name}Add(b *testing.B) {
	s := make([]{:name}, 1000)
	for i := range s {
		s[i] = rand{:name}()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%%len(s)]
		y := s[(i+1)%%len(s)]
		sink.{:name} = x.Add(y)
	}
}

func Benchmark{:name}Sub(b *testing.B) {
	s := make([]{:name}, 1000)
	for i := range s {
		s[i] = rand{:name}()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%%len(s)]
		y := s[(i+1)%%len(s)]
		sink.{:name} = x.Sub(y)
	}
}

func Benchmark{:name}Mul(b *testing.B) {
	s := make([]{:name}, 1000)
	for i := range s {
		s[i] = rand{:name}()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := s[i%%len(s)]
		y := s[(i+1)%%len(s)]
		sink.{:name} = x.Mul(y)
	}
}

func Benchmark{:name}QuoRem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink.{:name}, sink.{:name} = U{:bits}(uint64(i + 2)).QuoRem(U{:bits}(uint64(i + 1)))
	}
}

func Benchmark{:name}QuoRem64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink.{:name}, sink.uint64 = U{:bits}(uint64(i + 2)).quoRem64(uint64(i + 1))
	}
}
`)
}
