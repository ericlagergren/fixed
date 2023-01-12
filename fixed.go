// Package fixed implements fixed-size numeric types.
package fixed

//go:generate go run github.com/ericlagergren/fixed/internal/cmd/gen 256 512 1024 2048

// Uint is an unsigned integer.
type Uint[T any] interface {
	// BitLen returns the number of bits required to represent x.
	BitLen() int
	// LeadingZeros returns the number of leading zeros in x.
	LeadingZeros() int
	// IsZero repors whether x is zero.
	IsZero() bool
	// Cmp compares x and y and returns
	//
	//   - +1 if x > y
	//   - 0 if x == y
	//   - -1 if x < y
	Cmp(T) int
	// Add returns x+y.
	Add(T) T
	// Sub returns x-y.
	Sub(T) T
	// Mul returns x*y.
	Mul(T) T
	// QuoRem returns (q, r) such that
	//
	//	q = x/y
	//	r = x - y*q
	QuoRem(T) (q, r T)
	// And returns x&y.
	And(T) T
	// Or returns x|y.
	Or(T) T
	// Xor returns x^y.
	Xor(T) T
	// Lsh returns x<<n.
	Lsh(uint) T
	// Rsh returns x>>n.
	Rsh(uint) T
	// String returns the base-10 representation of x.
	String() string

	mulCheck64(uint64) (T, bool)
	addCheck64(uint64) (T, uint64)
	max() T
}

func abs(x int) uint {
	if x < 0 {
		return uint(-x)
	}
	return uint(x)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func cmp(x, y uint64) int {
	switch {
	case x > y:
		return +1
	case x < y:
		return -1
	default:
		return 0
	}
}

// low96 returns the low 96 bits in x.
func (x Uint256) low96() Uint96 {
	return Uint96{x.u0, uint32(x.u1)}
}
