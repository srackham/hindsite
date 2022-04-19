package slice

/*
Comparable slice.
Based on https://gobyexample.com/collection-functions
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

// Any returns true if one of the items in the slice satisfies the predicate f.
func (slice Slice[T]) Any(f func(T) bool) bool {
	for _, v := range slice {
		if f(v) {
			return true
		}
	}
	return false
}

// All returns true if all of the items in the slice satisfy the predicate f.
func (slice Slice[T]) All(f func(T) bool) bool {
	for _, v := range slice {
		if !f(v) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all items in the slice that satisfy the predicate f.
func (slice Slice[T]) Filter(f func(T) bool) Slice[T] {
	result := make([]T, 0)
	for _, v := range slice {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map returns a new slice containing the results of applying the function f to each item in the `from` slice.
func Map[T1 comparable, T2 comparable](from Slice[T1], f func(T1) T2) Slice[T2] {
	result := make(Slice[T2], len(from))
	for i, v := range from {
		result[i] = f(v)
	}
	return result
}
