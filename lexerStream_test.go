package govaluate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIfNext(t *testing.T) {
	stream := newLexerStream("abc & def")

	assert.False(t, stream.isNext("abc", 1))
	assert.True(t, stream.isNext("abc", 0))

	forward(stream, 3)

	assert.True(t, stream.isNext("abc", 3))
	assert.False(t, stream.isNext("def", 0))

	forward(stream, 1)
	assert.True(t, stream.isNext("&", 0))

	forward(stream, 2)

	assert.True(t, stream.isNext("def", 0))
	assert.False(t, stream.isNext("abc", 0))
	assert.True(t, stream.isNext("abc", 6))
}

func forward(stream *lexerStream, characters int) {
	for x := 0; x < characters; x++ {
		stream.readCharacter()
	}
}
