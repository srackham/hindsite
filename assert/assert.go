/*
	Assertions package.
*/

package assert

import (
	"regexp"
	"strings"
	"testing"
)

// PassIf fails and prints formatted message if not ok.
func PassIf(t *testing.T, ok bool, format string, args ...any) {
	t.Helper()
	if !ok {
		t.Errorf(format, args...)
	}
}

func Equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	PassIf(t, got == want, "got %v, want %v", want, got)
}

func NotEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	PassIf(t, got != want, "did not want %v", got)
}

func True(t *testing.T, got bool) {
	t.Helper()
	PassIf(t, got, "should be true")
}

func False(t *testing.T, got bool) {
	t.Helper()
	PassIf(t, !got, "should be false")
}

func EqualValues[T comparable](t *testing.T, want, got []T) {
	t.Helper()
	PassIf(t, len(got) == len(want), "got %v, want %v", want, got)
	for k := range got {
		PassIf(t, got[k] == want[k], "got %v, want %v", want, got)
	}
}

func Panics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		recover()
	}()
	f()
	t.Error("should have panicked")
}

func Contains(t *testing.T, s, substr string) {
	t.Helper()
	PassIf(t, strings.Contains(s, substr), "%q does not contain %q", s, substr)
}

func ContainsPattern(t *testing.T, s, pattern string) {
	t.Helper()
	matched, _ := regexp.MatchString(pattern, s)
	PassIf(t, matched, "%q does not contain pattern %q", s, pattern)
}
