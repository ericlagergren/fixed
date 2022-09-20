// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fixed

import (
	"math"
	"math/bits"
)

// q = ( x1 << _W + x0 - r)/y. m = floor(( _B^2 - 1 ) / d - _B). Requiring x1<y.
// An approximate reciprocal with a reference to "Improved Division by Invariant Integers
// (IEEE Transactions on Computers, 11 Jun. 2010)"
func divWW(x1, x0, y, m uint64) (q, r uint64) {
	s := bits.LeadingZeros64(y)
	if s != 0 {
		x1 = x1<<s | x0>>(64-s)
		x0 <<= s
		y <<= s
	}
	d := y
	// We know that
	//   m = ⎣(B^2-1)/d⎦-B
	//   ⎣(B^2-1)/d⎦ = m+B
	//   (B^2-1)/d = m+B+delta1    0 <= delta1 <= (d-1)/d
	//   B^2/d = m+B+delta2        0 <= delta2 <= 1
	// The quotient we're trying to compute is
	//   quotient = ⎣(x1*B+x0)/d⎦
	//            = ⎣(x1*B*(B^2/d)+x0*(B^2/d))/B^2⎦
	//            = ⎣(x1*B*(m+B+delta2)+x0*(m+B+delta2))/B^2⎦
	//            = ⎣(x1*m+x1*B+x0)/B + x0*m/B^2 + delta2*(x1*B+x0)/B^2⎦
	// The latter two terms of this three-term sum are between 0 and 1.
	// So we can compute just the first term, and we will be low by at most 2.
	t1, t0 := bits.Mul64(m, x1)
	_, c := bits.Add64(t0, x0, 0)
	t1, _ = bits.Add64(t1, x1, c)
	// The quotient is either t1, t1+1, or t1+2.
	// We'll try t1 and adjust if needed.
	qq := t1
	// compute remainder r=x-d*q.
	dq1, dq0 := bits.Mul64(d, qq)
	r0, b := bits.Sub64(x0, dq0, 0)
	r1, _ := bits.Sub64(x1, dq1, b)
	// The remainder we just computed is bounded above by B+d:
	// r = x1*B + x0 - d*q.
	//   = x1*B + x0 - d*⎣(x1*m+x1*B+x0)/B⎦
	//   = x1*B + x0 - d*((x1*m+x1*B+x0)/B-alpha)                                   0 <= alpha < 1
	//   = x1*B + x0 - x1*d/B*m                         - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*((B^2-1)/d-B-beta)        - x1*d - x0*d/B + d*alpha   0 <= beta < 1
	//   = x1*B + x0 - x1*B + x1/B + x1*d + x1*d/B*beta - x1*d - x0*d/B + d*alpha
	//   =        x0        + x1/B        + x1*d/B*beta        - x0*d/B + d*alpha
	//   = x0*(1-d/B) + x1*(1+d*beta)/B + d*alpha
	//   <  B*(1-d/B) +  d*B/B          + d          because x0<B (and 1-d/B>0), x1<d, 1+d*beta<=B, alpha<1
	//   =  B - d     +  d              + d
	//   = B+d
	// So r1 can only be 0 or 1. If r1 is 1, then we know q was too small.
	// Add 1 to q and subtract d from r. That guarantees that r is <B, so
	// we no longer need to keep track of r1.
	if r1 != 0 {
		qq++
		r0 -= d
	}
	// If the remainder is still too large, increment q one more time.
	if r0 >= d {
		qq++
		r0 -= d
	}
	return qq, r0 >> s
}

// reciprocal return the reciprocal of the divisor. rec = floor(( _B^2 - 1 ) / u - _B). u = d1 << nlz(d1).
func reciprocal(d1 uint64) uint64 {
	u := d1 << bits.LeadingZeros64(d1)
	x1 := ^u
	const x0 = math.MaxUint64
	rec, _ := bits.Div64(x1, x0, u) // (_B^2-1)/U-_B = (_B*(_M-C)+_M)/U
	return rec
}

// mulAddWWW returns x*y + c.
func mulAddWWW(x, y, c uint64) (z1, z0 uint64) {
	hi, lo := bits.Mul64(x, y)
	lo, c = bits.Add64(lo, c, 0)
	return hi + c, lo
}

// mulAddWWWW returns (x*y + v) + c.
func mulAddWWWW(x, y, v, c uint64) (z1, z0 uint64) {
	hi, lo := mulAddWWW(x, y, v)
	z, cc := bits.Add64(lo, c, 0)
	return hi + cc, z
}
