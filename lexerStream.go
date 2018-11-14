package govaluate

type lexerStream struct {
	source   []rune
	position int
	length   int
}

func newLexerStream(source string) *lexerStream {

	var ret *lexerStream
	var runes []rune

	for _, character := range source {
		runes = append(runes, character)
	}

	ret = new(lexerStream)
	ret.source = runes
	ret.length = len(runes)
	return ret
}

func (this *lexerStream) readCharacter() rune {

	var character rune

	character = this.source[this.position]
	this.position += 1
	return character
}

func (this *lexerStream) rewind(amount int) {
	this.position -= amount
}

func (this lexerStream) canRead() bool {
	return this.position < this.length
}

func (this lexerStream) isNext(clause string, rewind int) bool {
	if (this.length-this.position+rewind) >= len(clause) && this.position-rewind >= 0 {
		return clause == string(this.source[this.position-rewind:this.position+len(clause)-rewind])
	}

	return false
}

func (this *lexerStream) forward(count int) {
	if (this.length - this.position) >= count {
		this.position += count
	}
}
