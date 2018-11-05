govaluate
====

Evaluates logical expressions such way that you tell us what operators
and operands do, we just call them when the time comes, respecting precedences. 

Why can't you just write these expressions in code?
--

Sometimes, you can't know ahead-of-time what an expression will look like, or you want those expressions to be configurable.
Perhaps you've got a set of data running through your application, and you want to allow your users to specify some validations to run on it before committing it to a database. Or maybe you've written a monitoring framework which is capable of gathering a bunch of metrics, then evaluating a few expressions to see if any metrics should be alerted upon, but the conditions for alerting are different for each monitor.

How do I use it?
--

You create a new EvaluableExpression, then call "Evaluate" on it.

```
expression, err := NewEvaluableExpressionBuilder(func(operandName string) (interface{}, error) {
   		return "do something with operand", nil
   	}).
    WithOperator("&&", andOperator()).
    WithOperator("||", orOperator()).
    WithOperator("-", substractOperator()).
    Build("aaa || (bbb - ccc || (ddd && eee)) && fff")
expression.Evaluate()
```


License
--

This project is licensed under the MIT general use license. You're free to integrate, fork, and play with this code as 
you feel fit without consulting the author, as long as you provide proper credit to the author in your works.
This is fork of Knetic/govaluate project by George Lester.
