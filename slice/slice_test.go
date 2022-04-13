package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_set(t *testing.T) {
	assert := assert.New(t)
	s := Slice[string]{"foo", "bar"}
	assert.Equal(2, len(s))
	assert.Equal(1, s.IndexOf("bar"))
	assert.True(s.Has("foo"))
	assert.False(s.Has("qux"))
	s2 := Slice[int]{1, 2, 3, 4}
	assert.Equal(4, len(s2))
}
