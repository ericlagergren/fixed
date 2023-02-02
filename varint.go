package fixed

// MaxVarintLen returns the maximum number of bytes needed to
// represent a varint.
func MaxVarintLen[T Uint[T]]() int {
	return ((*new(T)).Size() + 7) / 7
}

// AppendUvarint appends the unsigned varint encoding of x to b
// and returns the resulting slice.
func AppendUvarint[T Uint[T]](b []byte, v T) []byte {
	for v.cmp64(0x80) >= 0 {
		b = append(b, v.uint8())
		v = v.Rsh(7)
	}
	return append(b, v.uint8())
}

// VarintLen returns the number of bytes required to encode x.
func VarintLen[T Uint[T]](v T) int {
	var n int
	for v.cmp64(0x80) >= 0 {
		n++
		v = v.Rsh(7)
	}
	return n + 1
}

// Uvarint parses an unsigned varint from b and returns that
// value and the number of bytes read.
func Uvarint[T Uint[T]](b []byte) (T, int) {
	bits := (*new(T)).Size()
	maxLen := (bits + 7) / 7
	maxVal := uint8((1 << (bits % 7)) - 1)

	var x T
	var s uint
	for i, c := range b {
		if i == maxLen {
			// Catch byte reads past maxLen.
			// See issue https://golang.org/issues/41185
			return *new(T), -(i + 1) // overflow
		}
		if c < 0x80 {
			if i == maxLen-1 && c > maxVal {
				return *new(T), -(i + 1) // overflow
			}
			return x.orLsh64(uint64(c), s), i + 1
		}
		x = x.orLsh64(uint64(c&0x7f), s)
		s += 7
	}
	return *new(T), 0
}
