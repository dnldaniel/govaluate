package govaluate

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"
)

func parseTokens(expression string, criterionFunction OperandHandler, operatorBySymbol map[string]EvaluationOperator) ([]ExpressionToken, error) {

	var ret []ExpressionToken
	var token ExpressionToken
	var stream *lexerStream
	var err error
	var found bool

	stream = newLexerStream(expression)

	for stream.canRead() {

		token, err, found = readToken(stream, criterionFunction, operatorBySymbol)

		if err != nil {
			return ret, err
		}

		if !found {
			break
		}

		// append this valid token
		ret = append(ret, token)
	}

	err = checkBalance(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func readToken(stream *lexerStream, criterionFunction OperandHandler, operatorBySymbol map[string]EvaluationOperator) (ExpressionToken, error, bool) {

	var ret ExpressionToken
	var tokenValue interface{}
	var tokenString string
	var kind TokenKind
	var character rune

	// numeric is 0-9, or . or 0x followed by digits
	// string starts with '
	// variable is alphanumeric, always starts with a letter
	// bracket always means variable
	// symbols are anything non-alphanumeric
	// all others read into a buffer until they reach the end of the stream
	for stream.canRead() {

		character = stream.readCharacter()

		if unicode.IsSpace(character) {
			continue
		}

		kind = UNKNOWN

		for setLogicalOperatorSymbol := range operatorBySymbol {
			if stream.isNext(setLogicalOperatorSymbol, 1) {
				kind = PROGRAMMABLE_OPERATOR
				tokenValue = setLogicalOperatorSymbol
				stream.forward(len(setLogicalOperatorSymbol) - 1)
				break
			}
		}

		if kind == PROGRAMMABLE_OPERATOR {
			break
		}

		// regular variable - or function?
		if unicode.IsLetter(character) {

			tokenString = readTokenUntilFalse(stream, isVariableName)

			kind = FUNCTION
			tokenValue = FunctionParameterPair{function: criterionFunction, parameter: tokenString}
			break
		}

		if character == '(' {
			tokenValue = character
			kind = CLAUSE
			break
		}

		if character == ')' {
			tokenValue = character
			kind = CLAUSE_CLOSE
			break
		}

		// must be a known symbol
		tokenString = readTokenUntilFalse(stream, isNotAlphanumeric)
		tokenValue = tokenString

		errorMessage := fmt.Sprintf("Invalid token: '%s'", tokenString)
		return ret, errors.New(errorMessage), false
	}

	ret.Kind = kind
	ret.Value = tokenValue

	return ret, nil, kind != UNKNOWN
}

func readTokenUntilFalse(stream *lexerStream, condition func(rune) bool) string {

	var ret string

	stream.rewind(1)
	ret, _ = readUntilFalse(stream, false, true, true, condition)
	return ret
}

/*
	Returns the string that was read until the given [condition] was false, or whitespace was broken.
	Returns false if the stream ended before whitespace was broken or condition was met.
*/
func readUntilFalse(stream *lexerStream, includeWhitespace bool, breakWhitespace bool, allowEscaping bool, condition func(rune) bool) (string, bool) {

	var tokenBuffer bytes.Buffer
	var character rune
	var conditioned bool

	conditioned = false

	for stream.canRead() {

		character = stream.readCharacter()

		// Use backslashes to escape anything
		if allowEscaping && character == '\\' {

			character = stream.readCharacter()
			tokenBuffer.WriteString(string(character))
			continue
		}

		if unicode.IsSpace(character) {

			if breakWhitespace && tokenBuffer.Len() > 0 {
				conditioned = true
				break
			}
			if !includeWhitespace {
				continue
			}
		}

		if condition(character) {
			tokenBuffer.WriteString(string(character))
		} else {
			conditioned = true
			stream.rewind(1)
			break
		}
	}

	return tokenBuffer.String(), conditioned
}

/*
	Checks to see if any optimizations can be performed on the given [tokens], which form a complete, valid expression.
	The returns slice will represent the optimized (or unmodified) list of tokens to use.
*/
func optimizeTokens(tokens []ExpressionToken) ([]ExpressionToken, error) {

	var token ExpressionToken
	var symbol OperatorSymbol
	var err error
	var index int

	for index, token = range tokens {

		// if we find a regex operator, and the right-hand value is a constant, precompile and replace with a pattern.
		if token.Kind != COMPARATOR {
			continue
		}

		symbol = comparatorSymbols[token.Value.(string)]
		if symbol != REQ && symbol != NREQ {
			continue
		}

		index++
		token = tokens[index]
		if token.Kind == STRING {

			token.Kind = PATTERN
			token.Value, err = regexp.Compile(token.Value.(string))

			if err != nil {
				return tokens, err
			}

			tokens[index] = token
		}
	}
	return tokens, nil
}

/*
	Checks the balance of tokens which have multiple parts, such as parenthesis.
*/
func checkBalance(tokens []ExpressionToken) error {

	var stream *tokenStream
	var token ExpressionToken
	var parens int

	stream = newTokenStream(tokens)

	for stream.hasNext() {

		token = stream.next()
		if token.Kind == CLAUSE {
			parens++
			continue
		}
		if token.Kind == CLAUSE_CLOSE {
			parens--
			continue
		}
	}

	if parens != 0 {
		return errors.New("Unbalanced parenthesis")
	}
	return nil
}

func isNotQuote(character rune) bool {

	return character != '\'' && character != '"'
}

func isNotAlphanumeric(character rune) bool {

	return !(unicode.IsDigit(character) ||
		unicode.IsLetter(character) ||
		character == '(' ||
		character == ')' ||
		character == '[' ||
		character == ']' || // starting to feel like there needs to be an `isOperation` func (#59)
		!isNotQuote(character))
}

func isVariableName(character rune) bool {

	return unicode.IsLetter(character) ||
		unicode.IsDigit(character) ||
		character == '_' ||
		character == '.'
}

/*
	Attempts to parse the [candidate] as a Time.
	Tries a series of standardized date formats, returns the Time if one applies,
	otherwise returns false through the second return.
*/
func tryParseTime(candidate string) (time.Time, bool) {

	var ret time.Time
	var found bool

	timeFormats := [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.Kitchen,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",                         // RFC 3339
		"2006-01-02 15:04",                   // RFC 3339 with minutes
		"2006-01-02 15:04:05",                // RFC 3339 with seconds
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
	}

	for _, format := range timeFormats {

		ret, found = tryParseExactTime(candidate, format)
		if found {
			return ret, true
		}
	}

	return time.Now(), false
}

func tryParseExactTime(candidate string, format string) (time.Time, bool) {

	var ret time.Time
	var err error

	ret, err = time.ParseInLocation(format, candidate, time.Local)
	if err != nil {
		return time.Now(), false
	}

	return ret, true
}
