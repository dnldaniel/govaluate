package govaluate

import (
	"errors"
	"fmt"
)

const shortCircuitHolder int = -1

/*
	EvaluableExpression represents a set of ExpressionTokens which, taken together,
	are an expression that can be evaluated down into a single value.
*/
type EvaluableExpression struct {
	/*
		Whether or not to safely check types when evaluating.
		If true, this library will return error messages when invalid types are used.
		If false, the library will panic when operators encounter types they can't use.

		This is exclusively for users who need to squeeze every ounce of speed out of the library as they can,
		and you should only set this to false if you know exactly what you're doing.
	*/
	ChecksTypes bool

	tokens           []ExpressionToken
	evaluationStages *evaluationStage
	inputExpression  string
}

type EvaluableExpressionBuilder struct {
	operandHandler   OperandHandler
	operatorBySymbol map[string]EvaluationOperator
}

func NewEvaluableExpressionBuilder(operandHandler OperandHandler) *EvaluableExpressionBuilder {
	return &EvaluableExpressionBuilder{operandHandler: operandHandler, operatorBySymbol: make(map[string]EvaluationOperator)}
}

func (eeb *EvaluableExpressionBuilder) WithOperator(symbol string, symbolHandler EvaluationOperator) *EvaluableExpressionBuilder {
	eeb.operatorBySymbol[symbol] = symbolHandler
	return eeb
}

func (eeb *EvaluableExpressionBuilder) Build(expression string) (*EvaluableExpression, error) {
	return NewEvaluableExpression(expression, eeb.operandHandler, eeb.operatorBySymbol)
}

func NewEvaluableExpression(expression string, function OperandHandler, operatorBySymbol map[string]EvaluationOperator) (*EvaluableExpression, error) {

	var ret *EvaluableExpression
	var err error

	ret = new(EvaluableExpression)
	ret.inputExpression = expression

	ret.tokens, err = parseTokens(expression, function, operatorBySymbol)
	if err != nil {
		return nil, err
	}

	err = checkBalance(ret.tokens)
	if err != nil {
		return nil, err
	}

	err = checkExpressionSyntax(ret.tokens)
	if err != nil {
		return nil, err
	}

	ret.tokens, err = optimizeTokens(ret.tokens)
	if err != nil {
		return nil, err
	}

	ret.evaluationStages, err = planStages(ret.tokens, operatorBySymbol)
	if err != nil {
		return nil, err
	}

	ret.ChecksTypes = true
	return ret, nil
}

/*
	Runs the entire expression, evaluating operands and then operators on top of operands.
*/
func (ee *EvaluableExpression) Evaluate() (interface{}, error) {
	if ee.evaluationStages == nil {
		return nil, nil
	}

	return ee.evaluateStage(ee.evaluationStages)
}

func (ee *EvaluableExpression) evaluateStage(stage *evaluationStage) (interface{}, error) {

	var left, right interface{}
	var err error

	if stage.leftStage != nil {
		left, err = ee.evaluateStage(stage.leftStage)
		if err != nil {
			return nil, err
		}
	}

	if right != shortCircuitHolder && stage.rightStage != nil {
		right, err = ee.evaluateStage(stage.rightStage,)
		if err != nil {
			return nil, err
		}
	}

	if ee.ChecksTypes {
		if stage.typeCheck == nil {

			err = typeCheck(stage.leftTypeCheck, left, stage.symbol, stage.typeErrorFormat)
			if err != nil {
				return nil, err
			}

			err = typeCheck(stage.rightTypeCheck, right, stage.symbol, stage.typeErrorFormat)
			if err != nil {
				return nil, err
			}
		} else {
			// special case where the type check needs to know both sides to determine if the operator can handle it
			if !stage.typeCheck(left, right) {
				errorMsg := fmt.Sprintf(stage.typeErrorFormat, left, stage.symbol.String())
				return nil, errors.New(errorMsg)
			}
		}
	}

	return stage.operator(left, right)
}

func typeCheck(check stageTypeCheck, value interface{}, symbol OperatorSymbol, format string) error {

	if check == nil {
		return nil
	}

	if check(value) {
		return nil
	}

	errorMsg := fmt.Sprintf(format, value, symbol.String())
	return errors.New(errorMsg)
}

/*
	Returns an array representing the ExpressionTokens that make up this expression.
*/
func (ee *EvaluableExpression) Tokens() []ExpressionToken {

	return ee.tokens
}

/*
	Returns the original expression used to create this EvaluableExpression.
*/
func (ee *EvaluableExpression) String() string {

	return ee.inputExpression
}
