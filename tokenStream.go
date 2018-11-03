package govaluate

type tokenStream struct {
	tokens      []ExpressionToken
	index       int
	tokenLength int
}

func newTokenStream(tokens []ExpressionToken) *tokenStream {

	var ret *tokenStream

	ret = new(tokenStream)
	ret.tokens = tokens
	ret.tokenLength = len(tokens)
	return ret
}

func (this *tokenStream) rewind() {
	this.index -= 1
}

func (this *tokenStream) next() ExpressionToken {

	var token ExpressionToken

	token = this.tokens[this.index]

	this.index += 1
	return token
}

func (this tokenStream) hasNext() bool {

	return this.index < this.tokenLength
}

func (this tokenStream) hasNextTokens(kinds ...TokenKind) bool {
	if len(kinds) > (this.tokenLength - this.index) {
		return false
	}

	index := this.index
	for k, v := range kinds {
		if v != this.tokens[index+k].Kind {
			return false
		}
	}

	return true
}
