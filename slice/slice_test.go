package slice

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	assert := assert.New(t)
	s := New("foo", "bar", "baz")
	assert.Equal("foo", s[0])
	assert.Equal("bar", s[1])
	assert.Equal(3, len(s))
	s = Slice[string]{"foo", "bar"}
	assert.Equal(2, len(s))
	assert.Equal(1, s.IndexOf("bar"))
	assert.True(s.Has("foo"))
	assert.False(s.Has("qux"))
	s2 := Slice[int]{1, 2, 3, 4}
	assert.Equal(4, len(s2))
	assert.True(s2.Equal(New(1, 2, 3, 4)))
	assert.False(s2.Equal(New(1, 2, 3)))
	assert.False(s2.Equal(New(1, 2, 3, 5)))
	assert.True(s2.Any(func(v int) bool { return v == 2 }))
	assert.False(s2.Any(func(v int) bool { return v == 42 }))
	assert.True(s2.All(func(v int) bool { return v < 42 }))
	assert.False(s2.All(func(v int) bool { return v < 2 }))
	assert.Equal(Slice[int]{1, 3, 4}, s2.Filter(func(v int) bool { return v != 2 }))
	assert.Equal(Slice[string]{"1", "4", "9", "16"}, Map(s2, func(v int) string { return fmt.Sprint(v * v) }))
}
