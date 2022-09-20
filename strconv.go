package fixed

func parseUint[T Uint[T]](s string, base int, expOK bool) (T, int, int, error) {
	const fnParseUint = "ParseUintX"

	if s == "" {
		return *new(T), 0, 0, syntaxError(fnParseUint, s)
	}

	s0 := s
	switch {
	case 2 <= base && base <= 36:
		// valid base; nothing to do
	case base == 0:
		// Look for octal, hex prefix.
		base = 10
		if s[0] == '0' {
			switch {
			case len(s) >= 3 && lower(s[1]) == 'b':
				base = 2
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'o':
				base = 8
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'x':
				base = 16
				s = s[2:]
			default:
				base = 8
				s = s[1:]
			}
		}
	default:
		return *new(T), 0, 0, baseError(fnParseUint, s0, base)
	}

	if expOK {
		switch base {
		case 2, 8, 10, 16:
			// OK
		default:
			return *new(T), 0, 0, baseError(fnParseUint, s0, base)
		}
	}

	var n T
	dotIdx := -1
	for i, c := range []byte(s) {
		var d byte
		switch {
		case c == '.' && expOK:
			if dotIdx > 0 {
				return *new(T), 0, 0, syntaxError(fnParseUint, s0)
			}
			dotIdx = i
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lower(c) && lower(c) <= 'z':
			d = lower(c) - 'a' + 10
		default:
			return *new(T), 0, 0, syntaxError(fnParseUint, s0)
		}

		if d >= byte(base) {
			if !expOK || (c != 'e' && c != 'E') {
				return *new(T), 0, 0, syntaxError(fnParseUint, s0)
			}
			return n, i, dotIdx, nil
		}

		var ok bool
		n, ok = n.mulCheck64(uint64(base))
		if !ok {
			// n*base overflows
			return (*new(T)).max(), 0, 0, rangeError(fnParseUint, s0)
		}

		var carry uint64
		n, carry = n.addCheck64(uint64(d))
		if carry != 0 {
			// n+d overflows
			return (*new(T)).max(), 0, 0, rangeError(fnParseUint, s0)
		}
	}
	return n, len(s), dotIdx, nil
}
