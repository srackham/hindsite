package slice

/*
Comparable slice.
*/

type Slice[T comparable] []T

func New[T comparable](v ...T) Slice[T] {
	return append(Slice[T]{}, v...)
}

// IndexOf returns the first index of `v`, or -1 if no match is found.
func (slice Slice[T]) IndexOf(t T) int {
	for i, v := range slice {
		if v == t {
			return i
		}
	}
	return -1
}

// Has returns `true` if `v` is in the slice.
func (slice Slice[T]) Has(v T) bool {
	return slice.IndexOf(v) >= 0
}

func (slice Slice[T]) Equal(s Slice[T]) bool {
	if len(slice) != len(s) {
		return false
	}
	for i := range slice {
		if slice[i] != s[i] {
			return false
		}
	}
	return true
}
