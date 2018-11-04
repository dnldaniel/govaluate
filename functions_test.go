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

	result, err := expression.Eval(nil)
	assert.NoError(t, err)
	assert.Equal(t, "aaa", result)
}

func Test_UnclosedBracketError(t *testing.T) {
	_, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		Build("(aaa AND bbb")

	assert.Equal(t, "Unbalanced parenthesis", err.Error())
}

func Test_AndRecognizedAsFunction_ErrorAsFunctionCannotBePrecededByFunction(t *testing.T) {
	_, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		Build("aaa AND bbb")

	assert.Error(t, err)
}

func Test_TwoOperandsSeparatedByCustomOr(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "OR", "aaa", "bbb")
	gomock.InOrder(call1)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("OR", orOperator(&mock)).
		Build("aaa OR bbb")

	assert.NoError(t, err)

	result, err := expression.Eval(nil)
	assert.NoError(t, err)
	assert.Equal(t, "(aaa OR bbb)", result)
}

func Test_LotsOfBrackets_AND_OR(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mock := MyMock{controller: controller}
	call1 := controller.RecordCall(&mock, myMockMethodName, "AND", "bbb", "ccc")
	call2 := controller.RecordCall(&mock, myMockMethodName, "AND", "ddd", "eee")
	call3 := controller.RecordCall(&mock, myMockMethodName, "OR", "(bbb AND ccc)", "(ddd AND eee)")
	call4 := controller.RecordCall(&mock, myMockMethodName, "OR", "aaa", "((bbb AND ccc) OR (ddd AND eee))")
	call5 := controller.RecordCall(&mock, myMockMethodName, "AND", "(aaa OR ((bbb AND ccc) OR (ddd AND eee)))", "fff")
	gomock.InOrder(call1, call2, call3, call4, call5)

	expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
		return operandName, nil
	}).
		WithOperator("AND", andOperator(&mock)).
		WithOperator("OR", orOperator(&mock)).
		Build("aaa OR (bbb AND ccc OR (ddd AND eee)) AND fff")

	assert.NoError(t, err)

	result, err := expression.Eval(nil)
	assert.NoError(t, err)
	assert.Equal(t, "((aaa OR ((bbb AND ccc) OR (ddd AND eee))) AND fff)", result)
}

const myMockMethodName = "OperatorCalled"

func orOperator(mock *MyMock) EvaluationOperator {
	return func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
		mock.OperatorCalled("OR", left, right)
		return "(" + left.(string) + " OR " + right.(string) + ")", nil
	}
}

func andOperator(mock *MyMock) EvaluationOperator {
	return func(left interface{}, right interface{}, parameters Parameters) (interface{}, error) {
		mock.OperatorCalled("AND", left, right)
		return "(" + left.(string) + " AND " + right.(string) + ")", nil
	}
}

type MyMock struct {
	controller *gomock.Controller
}

func (mm *MyMock) OperatorCalled(operator, leftOperand, rightOperand interface{}) {
	mm.controller.Call(mm, myMockMethodName, operator, leftOperand, rightOperand)
}
