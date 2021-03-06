package govaluate

import (
	"errors"
	"fmt"
)

var stageSymbolMap = map[OperatorSymbol]EvaluationOperator{
	EQ:             equalStage,
	NEQ:            notEqualStage,
	GT:             gtStage,
	LT:             ltStage,
	GTE:            gteStage,
	LTE:            lteStage,
	REQ:            regexStage,
	NREQ:           notRegexStage,
	IN:             inStage,
	BITWISE_OR:     bitwiseOrStage,
	BITWISE_AND:    bitwiseAndStage,
	BITWISE_XOR:    bitwiseXORStage,
	BITWISE_LSHIFT: leftShiftStage,
	BITWISE_RSHIFT: rightShiftStage,
	PLUS:           addStage,
	MULTIPLY:       multiplyStage,
	DIVIDE:         divideStage,
	MODULUS:        modulusStage,
	EXPONENT:       exponentStage,
	NEGATE:         negateStage,
	INVERT:         invertStage,
	BITWISE_NOT:    bitwiseNotStage,
	TERNARY_TRUE:   ternaryIfStage,
	TERNARY_FALSE:  ternaryElseStage,
	COALESCE:       ternaryElseStage,
	SEPARATE:       separatorStage,
	SET_AND:        setAndStage,
	SET_OR:         setOrStage,
	SET_MINUS:      setMinusStage,
}

/*
	A "precedent" is a function which will recursively parse new evaluateionStages from a given stream of tokens.
	It's called a `precedent` because it is expected to handle exactly what precedence of operator,
	and defer to other `precedent`s for other operators.
*/
type precedent func(stream *tokenStream, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error)

/*
	A convenience function for specifying the behavior of a `precedent`.
	Most `precedent` functions can be described by the same function, just with different type checks, symbols, and error formats.
	This struct is passed to `makePrecedentFromPlanner` to create a `precedent` function.
*/
type precedencePlanner struct {
	validSymbols map[string]OperatorSymbol
	validKinds   []TokenKind

	typeErrorFormat string

	next      precedent
	nextRight precedent
}

/*
	Given a planner, creates a function which will evaluate a specific precedence level of operators,
	and link it to other `precedent`s which recurse to parse other precedence levels.
*/
func makePrecedentFromPlanner(planner *precedencePlanner) precedent {

	var generated precedent
	var nextRight precedent

	generated = func(stream *tokenStream, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {
		return planPrecedenceLevel(
			stream,
			planner.typeErrorFormat,
			planner.validSymbols,
			planner.validKinds,
			nextRight,
			planner.next,
			operatorBySymbol,
		)
	}

	if planner.nextRight != nil {
		nextRight = planner.nextRight
	} else {
		nextRight = generated
	}

	return generated
}

/*
	Creates a `evaluationStageList` object which represents an execution plan (or tree)
	which is used to completely evaluate a set of tokens at evaluation-time.
	The three stages of evaluation can be thought of as parsing strings to tokens, then tokens to a stage list, then evaluation with parameters.
*/
func planStages(tokens []ExpressionToken, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	stream := newTokenStream(tokens)

	stage, err := planTokens(stream, operatorBySymbol)
	if err != nil {
		return nil, err
	}

	// while we're now fully-planned, we now need to re-order same-precedence operators.
	// this could probably be avoided with a different planning method
	reorderStages(stage)

	stage = elideLiterals(stage)
	return stage, nil
}

func planTokens(stream *tokenStream, evaluatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	if !stream.hasNext() {
		return nil, nil
	}

	// operator symbol is not relevant in case of programmatic evaluation operators
	operatorsBySymbol := make(map[string]OperatorSymbol)
	for symbol := range evaluatorBySymbol {
		operatorsBySymbol[symbol] = SET_AND
	}
	myNext := makePrecedentFromPlanner(&precedencePlanner{
		validSymbols:    operatorsBySymbol,
		validKinds:      []TokenKind{PROGRAMMABLE_OPERATOR},
		typeErrorFormat: logicalErrorFormat,
		next:            planFunction,
	})

	return myNext(stream, evaluatorBySymbol)
}

/*
	The most usual method of parsing an evaluation stage for a given precedence.
	Most stages use the same logic
*/
func planPrecedenceLevel(
	stream *tokenStream,
	typeErrorFormat string,
	validSymbols map[string]OperatorSymbol,
	validKinds []TokenKind,
	rightPrecedent precedent,
	leftPrecedent precedent, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	var token ExpressionToken
	var symbol OperatorSymbol
	var leftStage, rightStage *evaluationStage
	var checks typeChecks
	var err error
	var keyFound bool

	if leftPrecedent != nil {

		leftStage, err = leftPrecedent(stream, operatorBySymbol)
		if err != nil {
			return nil, err
		}
	}

	for stream.hasNext() {

		token = stream.next()

		if len(validKinds) > 0 {

			keyFound = false
			for _, kind := range validKinds {
				if kind == token.Kind {
					keyFound = true
					break
				}
			}

			if !keyFound {
				break
			}
		}

		if validSymbols != nil {

			if !isString(token.Value) {
				break
			}

			symbol, keyFound = validSymbols[token.Value.(string)]
			if !keyFound {
				break
			}
		}

		if rightPrecedent != nil {
			rightStage, err = rightPrecedent(stream, operatorBySymbol)
			if err != nil {
				return nil, err
			}
		}

		checks = findTypeChecks(symbol)

		var operator EvaluationOperator
		if isString(token.Value) && operatorBySymbol[token.Value.(string)] != nil {
			operator = operatorBySymbol[token.Value.(string)]
		} else {
			operator = stageSymbolMap[symbol]
		}

		return &evaluationStage{

			symbol:     symbol,
			leftStage:  leftStage,
			rightStage: rightStage,
			operator:   operator,

			leftTypeCheck:   checks.left,
			rightTypeCheck:  checks.right,
			typeCheck:       checks.combined,
			typeErrorFormat: typeErrorFormat,
		}, nil
	}

	stream.rewind()
	return leftStage, nil
}

/*
	A special case where functions need to be of higher precedence than values, and need a special wrapped execution stage operator.
*/
func planFunction(stream *tokenStream, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	var token ExpressionToken
	var rightStage *evaluationStage
	//var err error

	token = stream.next()

	if token.Kind != FUNCTION {
		stream.rewind()
		return planAccessor(stream, operatorBySymbol)
	}

	// would be needed only if functions had embedded values
	// rightStage, err = planAccessor(stream)
	// if err != nil {
	//	return nil, err
	// }

	if stream.hasNextTokens(CLAUSE, CLAUSE_CLOSE) {
		stream.next()
		stream.next()
	}

	rightStage = &evaluationStage{
		rightStage: nil,
		operator:   noopStageRight,
		symbol:     NOOP,
	}

	return &evaluationStage{

		symbol:          FUNCTIONAL,
		rightStage:      rightStage,
		operator:        makeFunctionStage(token.Value.(FunctionParameterPair)),
		typeErrorFormat: "Unable to run function '%v': %v",
	}, nil
}

func planAccessor(stream *tokenStream, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	var token, otherToken ExpressionToken
	var rightStage *evaluationStage
	var err error

	if !stream.hasNext() {
		return nil, nil
	}

	token = stream.next()

	if token.Kind != ACCESSOR {
		stream.rewind()
		return planValue(stream, operatorBySymbol)
	}

	// check if this is meant to be a function or a field.
	// fields have a clause next to them, functions do not.
	// if it's a function, parse the arguments. Otherwise leave the right stage null.
	if stream.hasNext() {

		otherToken = stream.next()
		if otherToken.Kind == CLAUSE {

			stream.rewind()

			rightStage, err = planTokens(stream, nil)
			if err != nil {
				return nil, err
			}
		} else {
			stream.rewind()
		}
	}

	return &evaluationStage{

		symbol:          ACCESS,
		rightStage:      rightStage,
		operator:        makeAccessorStage(token.Value.([]string)),
		typeErrorFormat: "Unable to access parameter field or method '%v': %v",
	}, nil
}

/*
	A truly special precedence function, this handles all the "lowest-case" errata of the process, including literals, parmeters,
	clauses, and prefixes.
*/
func planValue(stream *tokenStream, operatorBySymbol map[string]EvaluationOperator) (*evaluationStage, error) {

	var token ExpressionToken
	var symbol OperatorSymbol
	var ret *evaluationStage
	var operator EvaluationOperator
	var err error

	if !stream.hasNext() {
		return nil, nil
	}

	token = stream.next()

	switch token.Kind {

	case CLAUSE:

		ret, err = planTokens(stream, operatorBySymbol)
		if err != nil {
			return nil, err
		}

		// advance past the CLAUSE_CLOSE token. We know that it's a CLAUSE_CLOSE, because at parse-time we check for unbalanced parens.
		stream.next()

		// the stage we got represents all of the logic contained within the parens
		// but for technical reasons, we need to wrap this stage in a "noop" stage which breaks long chains of precedence.
		// see github #33.
		ret = &evaluationStage{
			rightStage: ret,
			operator:   noopStageRight,
			symbol:     NOOP,
		}

		return ret, nil

	case CLAUSE_CLOSE:

		// when functions have empty params, this will be hit. In this case, we don't have any evaluation stage to do,
		// so we just return nil so that the stage planner continues on its way.
		stream.rewind()
		return nil, nil
	}

	if operator == nil {
		errorMsg := fmt.Sprintf("Unable to plan token kind: '%s', value: '%v'", token.Kind.String(), token.Value)
		return nil, errors.New(errorMsg)
	}

	return &evaluationStage{
		symbol:   symbol,
		operator: operator,
	}, nil
}

/*
	Convenience function to pass a triplet of typechecks between `findTypeChecks` and `planPrecedenceLevel`.
	Each of these members may be nil, which indicates that type does not matter for that value.
*/
type typeChecks struct {
	left     stageTypeCheck
	right    stageTypeCheck
	combined stageCombinedTypeCheck
}

/*
	Maps a given [symbol] to a set of typechecks to be used during runtime.
*/
func findTypeChecks(symbol OperatorSymbol) typeChecks {

	switch symbol {
	case GT:
		fallthrough
	case LT:
		fallthrough
	case GTE:
		fallthrough
	case LTE:
		return typeChecks{
			combined: comparatorTypeCheck,
		}
	case REQ:
		fallthrough
	case NREQ:
		return typeChecks{
			left:  isString,
			right: isRegexOrString,
		}
	case IN:
		return typeChecks{
			right: isArray,
		}
	case BITWISE_LSHIFT:
		fallthrough
	case BITWISE_RSHIFT:
		fallthrough
	case BITWISE_OR:
		fallthrough
	case BITWISE_AND:
		fallthrough
	case BITWISE_XOR:
		return typeChecks{
			left:  isFloat64,
			right: isFloat64,
		}
	case PLUS:
		return typeChecks{
			combined: additionTypeCheck,
		}
	case MULTIPLY:
		fallthrough
	case DIVIDE:
		fallthrough
	case MODULUS:
		fallthrough
	case EXPONENT:
		return typeChecks{
			left:  isFloat64,
			right: isFloat64,
		}
	case NEGATE:
		return typeChecks{
			right: isFloat64,
		}
	case INVERT:
		return typeChecks{
			right: isBool,
		}
	case BITWISE_NOT:
		return typeChecks{
			right: isFloat64,
		}
	case TERNARY_TRUE:
		return typeChecks{
			left: isBool,
		}

	// unchecked cases
	case EQ:
		fallthrough
	case NEQ:
		return typeChecks{}
	case TERNARY_FALSE:
		fallthrough
	case COALESCE:
		fallthrough
	default:
		return typeChecks{}
	}
}

/*
	During stage planning, stages of equal precedence are parsed such that they'll be evaluated in reverse order.
	For commutative operators like "+" or "-", it's no big deal. But for order-specific operators, it ruins the expected result.
*/
func reorderStages(rootStage *evaluationStage) {

	// traverse every rightStage until we find multiples in a row of the same precedence.
	var identicalPrecedences []*evaluationStage
	var currentStage, nextStage *evaluationStage
	var precedence, currentPrecedence operatorPrecedence

	nextStage = rootStage
	precedence = findOperatorPrecedenceForSymbol(rootStage.symbol)

	for nextStage != nil {

		currentStage = nextStage
		nextStage = currentStage.rightStage

		// left depth first, since this entire method only looks for precedences down the right side of the tree
		if currentStage.leftStage != nil {
			reorderStages(currentStage.leftStage)
		}

		currentPrecedence = findOperatorPrecedenceForSymbol(currentStage.symbol)

		if currentPrecedence == precedence {
			identicalPrecedences = append(identicalPrecedences, currentStage)
			continue
		}

		// precedence break.
		// See how many in a row we had, and reorder if there's more than one.
		if len(identicalPrecedences) > 1 {
			mirrorStageSubtree(identicalPrecedences)
		}

		identicalPrecedences = []*evaluationStage{currentStage}
		precedence = currentPrecedence
	}

	if len(identicalPrecedences) > 1 {
		mirrorStageSubtree(identicalPrecedences)
	}
}

/*
	Performs a "mirror" on a subtree of stages.
	This mirror functionally inverts the order of execution for all members of the [stages] list.
	That list is assumed to be a root-to-leaf (ordered) list of evaluation stages, where each is a right-hand stage of the last.
*/
func mirrorStageSubtree(stages []*evaluationStage) {

	var rootStage, inverseStage, carryStage, frontStage *evaluationStage

	stagesLength := len(stages)

	// reverse all right/left
	for _, frontStage = range stages {

		carryStage = frontStage.rightStage
		frontStage.rightStage = frontStage.leftStage
		frontStage.leftStage = carryStage
	}

	// end left swaps with root right
	rootStage = stages[0]
	frontStage = stages[stagesLength-1]

	carryStage = frontStage.leftStage
	frontStage.leftStage = rootStage.rightStage
	rootStage.rightStage = carryStage

	// for all non-root non-end stages, right is swapped with inverse stage right in list
	for i := 0; i < (stagesLength-2)/2+1; i++ {

		frontStage = stages[i+1]
		inverseStage = stages[stagesLength-i-1]

		carryStage = frontStage.rightStage
		frontStage.rightStage = inverseStage.rightStage
		inverseStage.rightStage = carryStage
	}

	// swap all other information with inverse stages
	for i := 0; i < stagesLength/2; i++ {

		frontStage = stages[i]
		inverseStage = stages[stagesLength-i-1]
		frontStage.swapWith(inverseStage)
	}
}

/*
	Recurses through all operators in the entire tree, eliding operators where both sides are literals.
*/
func elideLiterals(root *evaluationStage) *evaluationStage {

	if root.leftStage != nil {
		root.leftStage = elideLiterals(root.leftStage)
	}

	if root.rightStage != nil {
		root.rightStage = elideLiterals(root.rightStage)
	}

	return elideStage(root)
}

/*
	Elides a specific stage, if possible.
	Returns the unmodified [root] stage if it cannot or should not be elided.
	Otherwise, returns a new stage representing the condensed value from the elided stages.
*/
func elideStage(root *evaluationStage) *evaluationStage {

	var leftValue, rightValue, result interface{}
	var err error

	// right side must be a non-nil value. Left side must be nil or a value.
	if root.rightStage == nil ||
		root.rightStage.symbol != LITERAL ||
		root.leftStage == nil ||
		root.leftStage.symbol != LITERAL {
		return root
	}

	// don't elide some operators
	switch root.symbol {
	case SEPARATE:
		fallthrough
	case IN:
		return root
	}

	// both sides are values, get their actual values.
	// errors should be near-impossible here. If we encounter them, just abort this optimization.
	leftValue, err = root.leftStage.operator(nil, nil)
	if err != nil {
		return root
	}

	rightValue, err = root.rightStage.operator(nil, nil)
	if err != nil {
		return root
	}

	// typcheck, since the grammar checker is a bit loose with which operator symbols go together.
	err = typeCheck(root.leftTypeCheck, leftValue, root.symbol, root.typeErrorFormat)
	if err != nil {
		return root
	}

	err = typeCheck(root.rightTypeCheck, rightValue, root.symbol, root.typeErrorFormat)
	if err != nil {
		return root
	}

	if root.typeCheck != nil && !root.typeCheck(leftValue, rightValue) {
		return root
	}

	// pre-calculate, and return a new stage representing the result.
	result, err = root.operator(leftValue, rightValue)
	if err != nil {
		return root
	}

	return &evaluationStage{
		symbol:   LITERAL,
		operator: makeLiteralStage(result),
	}
}
