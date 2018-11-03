package govaluate

import "testing"

func TestIfNext(t *testing.T) {
	stream := newLexerStream("abc & def")
	for x := 0; x < 6; x++ {
		println(string(stream.readCharacter()))
	}

	isNext := stream.isNext("def", 1)
	println(isNext)
}
