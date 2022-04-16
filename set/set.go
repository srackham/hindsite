package set

/*
Generic Set type.

Taken from
https://dbuddy.medium.com/implementing-set-data-structure-in-go-using-generics-4a967f823bfb
and updated to work with Go 1.18.

*/

type Set[T comparable] map[T]bool

// Constructor to create new set
// Example :-  New[int]() to create a int set
// New(string)() to create a string set
func New[T comparable](values ...T) Set[T] {
	s := make(Set[T])
	s.Add(values...)
	return s
}

// Add values to set
func (s Set[T]) Add(values ...T) {
	for _, value := range values {
		s[value] = true
	}
}

// Delete values from set
func (s Set[T]) Delete(values ...T) {
	for _, value := range values {
		delete(s, value)
	}
}

// Length of set
func (s Set[T]) Len() int {
	return len(s)
}

// Method to check if element exists in set
func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

// Iterate over set using a callback
func (s Set[T]) Iterate(it func(T)) {
	for v := range s {
		it(v)
	}
}

// Convert set to slice of values
func (s Set[T]) Values() []T {
	values := []T{}
	s.Iterate(func(value T) {
		values = append(values, value)
	})
	return values
}

// Clone set
func (s Set[T]) Clone() Set[T] {
	set := make(Set[T])
	set.Add(s.Values()...)
	return set
}

// Union of two sets
func (s Set[T]) Union(other Set[T]) Set[T] {
	set := s.Clone()
	set.Add(other.Values()...)
	return set
}

// Intersection of two sets
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	set := make(Set[T])
	s.Iterate(func(value T) {
		if other.Has(value) {
			set.Add(value)
		}
	})
	return set
}
