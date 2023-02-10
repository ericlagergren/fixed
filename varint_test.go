package fixed

import (
	"fmt"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"
	"golang.org/x/exp/rand"
)

func TestUvarint(t *testing.T) {
	testUvarint[Uint96](t)
	testUvarint[Uint128](t)
	testUvarint[Uint192](t)
	testUvarint[Uint256](t)
	testUvarint[Uint512](t)
	testUvarint[Uint1024](t)
	testUvarint[Uint2048](t)
}

func testUvarint[T Uint[T]](t *testing.T) {
	t.Run(fmt.Sprintf("%T", *new(T)), func(t *testing.T) {
		max := MaxVarintLen[T]()
		b := make([]byte, max)

		var want T
		for j := 0; j < 10_000; j++ {
			want = want.add64(rand.Uint64())
			want = want.mul64(rand.Uint64())
			b = AppendUvarint(b[:0], want)
			if got := VarintLen(want); got != len(b) {
				t.Fatalf("got %d, expected %d", got, len(b))
			}
			got, n := Uvarint[T](b)
			if n != len(b) {
				t.Fatalf("got %d, expected %d", n, len(b))
			}
			if !gcmp.Equal(want, got) {
				t.Fatalf("%s", gcmp.Diff(want, got))
			}
		}
	})
}

func BenchmarkAppendUvarint(b *testing.B) {
	benchmarkAppendUvarint[Uint96](b)
	benchmarkAppendUvarint[Uint128](b)
	benchmarkAppendUvarint[Uint192](b)
	benchmarkAppendUvarint[Uint256](b)
	benchmarkAppendUvarint[Uint512](b)
	benchmarkAppendUvarint[Uint1024](b)
	benchmarkAppendUvarint[Uint2048](b)
}

func benchmarkAppendUvarint[T Uint[T]](b *testing.B) {
	b.Run(fmt.Sprintf("%T", *new(T)), func(b *testing.B) {
		max := MaxVarintLen[T]()
		buf := make([]byte, max)
		var want T
		for i := 0; i < 64; i++ {
			want = want.add64(rand.Uint64())
			want = want.mul64(rand.Uint64())
		}
		b.SetBytes(int64(len(buf)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf = AppendUvarint(buf[:0], want)
		}
	})
}
