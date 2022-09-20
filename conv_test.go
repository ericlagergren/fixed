package fixed

import "testing"

func TestDigits(t *testing.T) {
	max := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}
	v := uint64(1)
	for i := 1; i < 21; i++ {
		// 9, 99, 999, ...
		if n := digits(v - 1); n != max(i-1, 1) {
			t.Fatalf("digits(%d): expected %d, got %d", v, max(i-1, 1), n)
		}
		// 10, 100, 1000, ...
		if n := digits(v); n != i {
			t.Fatalf("digits(%d): expected %d, got %d", v, i, n)
		}
		// 11, 101, 1001, ...
		if n := digits(v + 1); n != i {
			t.Fatalf("digits(%d): expected %d, got %d", v, i, n)
		}
		v *= 10
	}
}
