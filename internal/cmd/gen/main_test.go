package main

import (
	"strings"
	"testing"
)

func sprintf(format string, args ...any) string {
	var b strings.Builder
	fprintf(&b, format, args...)
	return b.String()
}

func TestFprintf(t *testing.T) {
	for _, tc := range []struct {
		want   string
		format string
		args   []any
	}{
		{
			"hello, bob!",
			"hello, {:name}!",
			[]any{named("name", "bob")},
		},
		{
			"hello, bob! i'm 42",
			"hello, %s! i'm {:age}",
			[]any{"bob", named("age", 42)},
		},
		{
			"1 2 3 4 5 6",
			"%d %d %d {:a} {:b} {:c}",
			[]any{1, 2, 3, named("a", 4), named("b", 5), named("c", 6)},
		},
	} {
		got := sprintf(tc.format, tc.args...)
		if got != tc.want {
			t.Fatalf("expected %q, got %q", tc.want, got)
		}
	}
}
