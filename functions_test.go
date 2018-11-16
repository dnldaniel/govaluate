package govaluate

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_OperandOnly(t *testing.T) {
	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		Build("aaa")

	assert.NoError(t, err)

	result, err := expression.Evaluate()
	assert.NoError(t, err)
	assert.Equal(t, "aaa", result)
}

func Test_UnclosedBracketError(t *testing.T) {
	_, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("&&", andOperator(nil)).
		Build("(aaa && bbb")

	assert.Error(t, err)
	assert.Equal(t, "Unbalanced parenthesis", err.Error())
}

func Test_AndRecognizedAsFunction_ErrorAsFunctionCannotBePrecededByFunction(t *testing.T) {
	_, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		Build("aaa && bbb")

	assert.Error(t, err)
	assert.Equal(t, "Invalid token: '&&'", err.Error())
}

func Test_TwoOperandsSeparatedByCustomOr(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "||", "a22aa", "bbb")
	gomock.InOrder(call1)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("||", orOperator(&mock)).
		Build("a22aa || bbb")

	assert.NoError(t, err)

	result, err := expression.Evaluate()
	assert.NoError(t, err)
	assert.Equal(t, "(a22aa || bbb)", result)
}

func Test_TwoOperandsSeparatedBySomeMadeUpOperand(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "||", "a22aa", "bbb")
	call2 := controller.RecordCall(&mock, myMockMethodName, "$$", "(a22aa || bbb)", "ccc")
	gomock.InOrder(call1, call2)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("||", orOperator(&mock)).
		WithOperator("$$", anyOperator(&mock, "$$")).
		Build("a22aa || bbb $$ ccc")

	assert.NoError(t, err)

	result, err := expression.Evaluate()
	assert.NoError(t, err)
	assert.Equal(t, "((a22aa || bbb) $$ ccc)", result)
}

func Test_TwoOperandsSeparatedByCustomOrWithSpaceOrWithout(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "||", "aaa", "bbb")
	gomock.InOrder(call1)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("||", orOperator(&mock)).
		Build("aaa ||bbb")

	assert.NoError(t, err)

	result, err := expression.Evaluate()
	assert.NoError(t, err)
	assert.Equal(t, "(aaa || bbb)", result)
}

func Test_LotsOfBrackets_OPERATORS_AND_OR_SUBTRACT(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "-", "bbb", "ccc")
	call2 := controller.RecordCall(&mock, myMockMethodName, "&&", "ddd", "eee")
	call3 := controller.RecordCall(&mock, myMockMethodName, "||", "(bbb - ccc)", "(ddd && eee)")
	call4 := controller.RecordCall(&mock, myMockMethodName, "||", "aaa", "((bbb - ccc) || (ddd && eee))")
	call5 := controller.RecordCall(&mock, myMockMethodName, "&&", "(aaa || ((bbb - ccc) || (ddd && eee)))", "fff")
	gomock.InOrder(call1, call2, call3, call4, call5)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("&&", andOperator(&mock)).
		WithOperator("||", orOperator(&mock)).
		WithOperator("-", substractOperator(&mock)).
		Build("aaa || (bbb - ccc || (ddd && eee)) && fff")

	assert.NoError(t, err)

	result, err := expression.Evaluate()
	assert.NoError(t, err)
	assert.Equal(t, "((aaa || ((bbb - ccc) || (ddd && eee))) && fff)", result)
}

func orOperator(mock *MyMock) EvaluationOperator {
	return anyOperator(mock, "||")
}

func andOperator(mock *MyMock) EvaluationOperator {
	return anyOperator(mock, "&&")
}

func substractOperator(mock *MyMock) EvaluationOperator {
	return anyOperator(mock, "-")
}

func anyOperator(mock *MyMock, operatorSymbol string) EvaluationOperator {
	return func(left interface{}, right interface{}) (interface{}, error) {
		mock.OperatorCalled(operatorSymbol, left, right)
		return "(" + left.(string) + " " + operatorSymbol + " " + right.(string) + ")", nil
	}
}

const myMockMethodName = "OperatorCalled"

type MyMock struct {
	controller *gomock.Controller
}

func (mm *MyMock) OperatorCalled(operator, leftOperand, rightOperand interface{}) {
	mm.controller.Call(mm, myMockMethodName, operator, leftOperand, rightOperand)
}
