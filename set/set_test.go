package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	assert := assert.New(t)
	set1 := New[int]()
	assert.Equal(0, len(set1))
	set1.Add(1, 2, 3, 4, 2, 4)
	assert.Equal(4, len(set1))
	assert.Equal(1, set1.Count(1))
	assert.Equal(2, set1.Count(4))
	assert.Equal(0, set1.Count(42))
	assert.True(set1.Has(3))
	assert.False(set1.Has(0))
	set2 := New(3, 4, 5, 6, 7, 7)
	assert.Equal(5, len(set2))
	set3 := set1.Union(set2)
	assert.Equal(7, len(set3))
	set4 := set1.Intersection(set2)
	assert.Equal(2, len(set4))
	set5 := New("foo", "bar", "baz", "baz")
	assert.Equal(3, set5.Len())
	assert.True(set5.Has("foo"))
	set5.Delete("foo")
	assert.Equal(2, set5.Len())
	assert.False(set5.Has("foo"))
}
