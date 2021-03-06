package govaluate

/*
	Represents the valid symbols for operators.

*/
type OperatorSymbol int

const (
	VALUE OperatorSymbol = iota
	LITERAL
	NOOP
	EQ
	NEQ
	GT
	LT
	GTE
	LTE
	REQ
	NREQ
	IN

	SET_AND
	SET_OR
	SET_MINUS

	PLUS
	BITWISE_AND
	BITWISE_OR
	BITWISE_XOR
	BITWISE_LSHIFT
	BITWISE_RSHIFT
	MULTIPLY
	DIVIDE
	MODULUS
	EXPONENT

	NEGATE
	INVERT
	BITWISE_NOT

	TERNARY_TRUE
	TERNARY_FALSE
	COALESCE

	FUNCTIONAL
	ACCESS
	SEPARATE
)

type operatorPrecedence int

const (
	noopPrecedence operatorPrecedence = iota
	valuePrecedence
	functionalPrecedence
	prefixPrecedence
	exponentialPrecedence
	bitwisePrecedence
	bitwiseShiftPrecedence
	multiplicativePrecedence
	comparatorPrecedence
	ternaryPrecedence
	logicalAndPrecedence
	separatePrecedence
)

func findOperatorPrecedenceForSymbol(symbol OperatorSymbol) operatorPrecedence {

	switch symbol {
	case NOOP:
		return noopPrecedence
	case VALUE:
		return valuePrecedence
	case EQ:
		fallthrough
	case NEQ:
		fallthrough
	case GT:
		fallthrough
	case LT:
		fallthrough
	case GTE:
		fallthrough
	case LTE:
		fallthrough
	case REQ:
		fallthrough
	case NREQ:
		fallthrough
	case IN:
		return comparatorPrecedence
	case SET_AND:
		return logicalAndPrecedence
	case SET_OR:
		return logicalAndPrecedence
	case SET_MINUS:
		return logicalAndPrecedence
	case BITWISE_AND:
		fallthrough
	case BITWISE_OR:
		fallthrough
	case BITWISE_XOR:
		return bitwisePrecedence
	case BITWISE_LSHIFT:
		fallthrough
	case BITWISE_RSHIFT:
		return bitwiseShiftPrecedence
	case PLUS:
		fallthrough
	case MULTIPLY:
		fallthrough
	case DIVIDE:
		fallthrough
	case MODULUS:
		return multiplicativePrecedence
	case EXPONENT:
		return exponentialPrecedence
	case BITWISE_NOT:
		fallthrough
	case NEGATE:
		fallthrough
	case INVERT:
		return prefixPrecedence
	case COALESCE:
		fallthrough
	case TERNARY_TRUE:
		fallthrough
	case TERNARY_FALSE:
		return ternaryPrecedence
	case ACCESS:
		fallthrough
	case FUNCTIONAL:
		return functionalPrecedence
	case SEPARATE:
		return separatePrecedence
	}

	return valuePrecedence
}

/*
	Map of all valid comparators, and their string equivalents.
	Used during parsing of expressions to determine if a symbol is, in fact, a comparator.
	Also used during evaluation to determine exactly which comparator is being used.
*/
var comparatorSymbols = map[string]OperatorSymbol{
	"==": EQ,
	"!=": NEQ,
	">":  GT,
	">=": GTE,
	"<":  LT,
	"<=": LTE,
	"=~": REQ,
	"!~": NREQ,
	"in": IN,
}

var ternarySymbols = map[string]OperatorSymbol{
	"?":  TERNARY_TRUE,
	":":  TERNARY_FALSE,
	"??": COALESCE,
}

/*
	Returns true if this operator is contained by the given array of candidate symbols.
	False otherwise.
*/
func (this OperatorSymbol) IsModifierType(candidate []OperatorSymbol) bool {

	for _, symbolType := range candidate {
		if this == symbolType {
			return true
		}
	}

	return false
}

/*
	Generally used when formatting type check errors.
	We could store the stringified symbol somewhere else and not require a duplicated codeblock to translate
	OperatorSymbol to string, but that would require more memory, and another field somewhere.
	Adding operators is rare enough that we just stringify it here instead.
*/
func (this OperatorSymbol) String() string {

	switch this {
	case NOOP:
		return "NOOP"
	case VALUE:
		return "VALUE"
	case EQ:
		return "="
	case NEQ:
		return "!="
	case GT:
		return ">"
	case LT:
		return "<"
	case GTE:
		return ">="
	case LTE:
		return "<="
	case REQ:
		return "=~"
	case NREQ:
		return "!~"
	case SET_AND:
		return "&&"
	case SET_OR:
		return "||"
	case SET_MINUS:
		return "-"
	case IN:
		return "in"
	case BITWISE_AND:
		return "&"
	case BITWISE_OR:
		return "|"
	case BITWISE_XOR:
		return "^"
	case BITWISE_LSHIFT:
		return "<<"
	case BITWISE_RSHIFT:
		return ">>"
	case PLUS:
		return "+"
	case MULTIPLY:
		return "*"
	case DIVIDE:
		return "/"
	case MODULUS:
		return "%"
	case EXPONENT:
		return "**"
	case INVERT:
		return "!"
	case BITWISE_NOT:
		return "~"
	case TERNARY_TRUE:
		return "?"
	case TERNARY_FALSE:
		return ":"
	case COALESCE:
		return "??"
	}
	return ""
}
