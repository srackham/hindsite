package slice

import (
	"fmt"
	"testing"

	"github.com/srackham/hindsite/v2/assert"
)

func TestSlice(t *testing.T) {
	s := New("foo", "bar", "baz")
	assert.Equal(t, "foo", s[0])
	assert.Equal(t, "bar", s[1])
	assert.Equal(t, 3, len(s))
	s = Slice[string]{"foo", "bar"}
	assert.Equal(t, 2, len(s))
	assert.Equal(t, 1, s.IndexOf("bar"))
	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("qux"))
	s2 := Slice[int]{1, 2, 3, 4}
	assert.Equal(t, 4, len(s2))
	assert.True(t, s2.Equal(New(1, 2, 3, 4)))
	assert.False(t, s2.Equal(New(1, 2, 3)))
	assert.False(t, s2.Equal(New(1, 2, 3, 5)))
	assert.True(t, s2.Any(func(v int) bool { return v == 2 }))
	assert.False(t, s2.Any(func(v int) bool { return v == 42 }))
	assert.True(t, s2.All(func(v int) bool { return v < 42 }))
	assert.False(t, s2.All(func(v int) bool { return v < 2 }))
	assert.EqualValues(t, Slice[int]{1, 3, 4}, s2.Filter(func(v int) bool { return v != 2 }))
	assert.EqualValues(t, Slice[string]{"1", "4", "9", "16"}, Map(s2, func(v int) string { return fmt.Sprint(v * v) }))
}
