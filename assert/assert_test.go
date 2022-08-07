package assert

import (
	"testing"
)

func TestAssertFunctions(t *testing.T) {
	Equal(t, 1, 1)
	NotEqual(t, 1, 2)
	Equal(t, "hello", "hello")
	True(t, true)
	False(t, false)
	EqualValues(t, []string{"one", "two"}, []string{"one", "two"})
	Panics(t, func() { panic("panics test") })
	var p *int
	Equal(t, p, (*int)(nil))
	Equal(t, p, nil)
	True(t, p == nil)
	PassIf(t, p == nil, "p should be nil")
	n := 42
	p = &n
	True(t, p != nil)
	Contains(t, "foobar", "bar")

	/* Uncomment to see failures: */
	// PassIf(t, false, "the criteria is not met")
	// Equal(t, "hello", "Grace")
	// True(t, false)
	// EqualValues(t, []string{"one", "two"}, []string{"one"})
	// EqualValues(t, []string{"one", "two"}, []string{"one", "two", "three"})
	// EqualValues(t, []string{"one", "two"}, []string{})
	// Panics(t, func() {})
	// Contains(t, "foobar", "\"baz")

	/* Uncomment to see the compilation error: */
	// Equal(t, 1, "1")
}
