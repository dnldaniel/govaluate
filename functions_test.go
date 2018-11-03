package govaluate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	inputExpression := "aaa OR (bbb AND ccc OR (ddd AND eee))"
	println(inputExpression)


	var criterionHandler OperandHandler

	criterionHandler = func(criterionName string) (interface{}, error) {
		println("clientSideFunction called: " + criterionName)
		return criterionName, nil
	}

	expression, err := NewEvaluableExpressionBuilder(criterionHandler).WithOperator("AND", setAndStage).Build("aaa OR (bbb AND ccc OR (ddd AND eee))")

	result, err := expression.Eval(MapParameters(map[string]interface{}{}))
	println(result.(string))

	assert.NoError(t, err)
}

func Test2(t *testing.T) {

}
