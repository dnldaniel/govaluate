package govaluate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	inputExpression := "aaa OR (bbb AND ccc OR (ddd AND eee))"
	println(inputExpression)

	criterionHandler := func(criterionName string) (interface{}, error) {
		println("clientSideFunction called: " + criterionName)
		return criterionName, nil
	}
	var setOrHandler EvaluationOperator = func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
		return "whatever", nil
	}

	evaluableExpression, err := NewFunctionalCriteriaExpression(inputExpression,
		criterionHandler).OverrideOperator("OR", setOrHandler)
	result, err := evaluableExpression.Eval(MapParameters(map[string]interface{}{}))
	println(result.(string))

	assert.NoError(t, err)

	println(evaluableExpression)
	println(err)
}

func Test2(t *testing.T) {

}
