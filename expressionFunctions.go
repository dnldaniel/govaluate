package govaluate

/*
	Represents a function that can be called from within an expression.
	This method must return an error if, for any reason, it is unable to produce exactly one unambiguous result.
	An error returned will halt execution of the expression.
*/
type OperandHandler func(operandName string) (interface{}, error)

type FunctionParameterPair struct {
	function  OperandHandler
	parameter string
}

func (fpp *FunctionParameterPair) NewFunction() (interface{}, error) {
	return fpp.function(fpp.parameter)
}
