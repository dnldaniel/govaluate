package govaluate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	inputExpression := "aa_a OR (bbb AND ccc OR (ddd AND eee))"
	println(inputExpression)

	var criterionHandler OperandHandler

	criterionHandler = func(criterionName string) (interface{}, error) {
		println("clientSideFunction called: " + criterionName)
		return criterionName, nil
	}


	dupa := func (left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
	return "(" + left.(string) + " DUPA " + right.(string) + ")", nil
	}

	expression, err := NewEvaluableExpressionBuilder(criterionHandler).
		WithOperator("AND", dupa).
		WithOperator("OR", setOrStage).
		Build(inputExpression)

	assert.NoError(t, err)

	result, err := expression.Eval(MapParameters(map[string]interface{}{}))
	println(result.(string))

	assert.NoError(t, err)
}

func Test2(t *testing.T) {

}
