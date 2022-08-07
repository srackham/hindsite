package set

import (
	"testing"

	"github.com/srackham/hindsite/v2/assert"
)

func TestSet(t *testing.T) {
	set1 := New[int]()
	assert.Equal(t, 0, len(set1))
	set1.Add(1, 2, 3, 4, 2, 4)
	assert.Equal(t, 4, len(set1))
	assert.Equal(t, 1, set1.Count(1))
	assert.Equal(t, 2, set1.Count(4))
	assert.Equal(t, 0, set1.Count(42))
	assert.True(t, set1.Has(3))
	assert.False(t, set1.Has(0))
	set2 := New(3, 4, 5, 6, 7, 7)
	assert.Equal(t, 5, len(set2))
	set3 := set1.Union(set2)
	assert.Equal(t, 7, len(set3))
	set4 := set1.Intersection(set2)
	assert.Equal(t, 2, len(set4))
	set5 := New("foo", "bar", "baz", "baz")
	assert.Equal(t, 3, set5.Len())
	assert.True(t, set5.Has("foo"))
	set5.Delete("foo")
	assert.Equal(t, 2, set5.Len())
	assert.False(t, set5.Has("foo"))
}
