package mylang

import (
	"fmt"
	stdlog "log"
	"os"
)

var (
	Logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags)
)

func SetLogger(logger *stdlog.Logger) {
	Logger = logger
}

// Environment 代表执行环境
type Environment struct {
	store map[string]interface{}
	outer *Environment
}

// NewEnvironment 创建一个新的执行环境
func NewEnvironment() *Environment {
	s := make(map[string]interface{})
	return &Environment{store: s}
}

// Get 从环境中获取一个值
func (e *Environment) Get(name string) (interface{}, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		val, ok = e.outer.Get(name)
	}
	return val, ok
}

// Set 在环境中设置一个值
func (e *Environment) Set(name string, val interface{}) interface{} {
	e.store[name] = val
	return val
}

// Interpreter 代表解释器
type Interpreter struct {
	env         *Environment
	CustomVariableGetter func (name string) (any) // if nil, use env.Get(name) to get the variable value
	drawingVars map[string]struct{} // 记录画图变量
	Err         error
}

// NewInterpreter 创建一个新的解释器
func NewInterpreter(env *Environment) *Interpreter {
	return &Interpreter{
		env:         env,
		drawingVars: make(map[string]struct{}),
	}
}

// Eval 评估一个节点
func (i *Interpreter) Eval(node Node) interface{} {
	Logger.Println("Evaluating node of type", fmt.Sprintf("%T", node))
	switch node := node.(type) {
	case *Program:
		return i.evalProgram(node)
	case *AssignmentStatement:
		return i.evalAssignmentStatement(node)
	case *ExpressionStatement:
		return i.Eval(node.Expression)
	case *BinaryExpression:
		Logger.Println("Evaluating binary expression")
		return i.evalBinaryExpression(node)
	case *NumberLiteral:
		Logger.Println("Evaluating number literal:", node.Value)
		return node.Value
	case *Identifier:
		Logger.Println("Evaluating identifier:", node.Value)
		val := i.evalIdentifier(node)
		if _, ok := val.(func([]interface{}) interface{}); ok {
			Logger.Println("Identifier", node.Value, "is a function")
			return val
		}
		return val
	case *FunctionCall:
		Logger.Println("Evaluating function call")
		return i.evalFunctionCall(node)
	}
	return nil
}

func (i *Interpreter) evalProgram(program *Program) interface{} {
	var result interface{}
	Logger.Println("Evaluating program with", len(program.Statements), "statements")
	for idx, statement := range program.Statements {
		Logger.Println("Evaluating statement", idx)
		result = i.Eval(statement)
		Logger.Println("Result of statement", idx, ":", result)
	}
	return result
}

func (i *Interpreter) evalAssignmentStatement(stmt *AssignmentStatement) interface{} {
	if stmt == nil || stmt.Name == nil {
		Logger.Println("Assignment statement or name is nil")
		return nil
	}
	val := i.Eval(stmt.Value)
	Logger.Println("Setting variable", stmt.Name.Value, "to", val)

	// 如果是画图变量赋值，记录到画图变量映射中
	if stmt.IsDrawingVar {
		i.drawingVars[stmt.Name.Value] = struct{}{}
		Logger.Println("Added", stmt.Name.Value, "to drawing variables")
	}

	return i.env.Set(stmt.Name.Value, val)
}

func (i *Interpreter) evalBinaryExpression(be *BinaryExpression) interface{} {
	left := i.Eval(be.Left)
	right := i.Eval(be.Right)
	Logger.Println("Binary expression, left:", left, "operator:", be.Operator, "right:", right)

	// 处理浮点数
	leftVal, lok := left.(float64)
	rightVal, rok := right.(float64)

	if lok && rok {
		switch be.Operator {
		case "+":
			return leftVal + rightVal
		case "-":
			return leftVal - rightVal
		case "*":
			return leftVal * rightVal
		case "/":
			if rightVal == 0 {
				return nil
			}
			return leftVal / rightVal
		}
	}

	// 处理数组和浮点数的操作
	leftArr, larrOk := left.([]float64)
	rightArr, rarrOk := right.([]float64)

	if larrOk && rok { // 数组与浮点数操作
		result := make([]float64, len(leftArr))
		for i := 0; i < len(leftArr); i++ {
			switch be.Operator {
			case "+":
				result[i] = leftArr[i] + rightVal
			case "-":
				result[i] = leftArr[i] - rightVal
			case "*":
				result[i] = leftArr[i] * rightVal
			case "/":
				if rightVal == 0 {
					return nil
				}
				result[i] = leftArr[i] / rightVal
			}
		}
		return result
	} else if lok && rarrOk { // 浮点数与数组操作
		result := make([]float64, len(rightArr))
		for i := 0; i < len(rightArr); i++ {
			switch be.Operator {
			case "+":
				result[i] = leftVal + rightArr[i]
			case "-":
				result[i] = leftVal - rightArr[i]
			case "*":
				result[i] = leftVal * rightArr[i]
			case "/":
				result[i] = leftVal / rightArr[i]
			}
		}
		return result
	} else if larrOk && rarrOk { // 数组与数组操作
		if len(leftArr) != len(rightArr) {
			return nil
		}
		result := make([]float64, len(leftArr))
		for i := 0; i < len(leftArr); i++ {
			switch be.Operator {
			case "+":
				result[i] = leftArr[i] + rightArr[i]
			case "-":
				result[i] = leftArr[i] - rightArr[i]
			case "*":
				result[i] = leftArr[i] * rightArr[i]
			case "/":
				if rightArr[i] == 0 {
					return nil
				}
				result[i] = leftArr[i] / rightArr[i]
			}
		}
		return result
	}
	return nil
}

func (i *Interpreter) evalIdentifier(symbol *Identifier) interface{} {
	if i.CustomVariableGetter != nil {
		x := i.CustomVariableGetter(symbol.Value)
		if x != nil {
			return x
		}
	}
	if val, ok := i.env.Get(symbol.Value); ok {
		Logger.Println("Found identifier", symbol.Value, "with value", val)
		return val
	}
	Logger.Println("Identifier", symbol.Value, "not found")
	return nil
}

func (i *Interpreter) evalFunctionCall(fc *FunctionCall) interface{} {
	defer func() {
		err := recover()
		if err != nil {
			Logger.Printf("panic: %v", err)
			i.Err = fmt.Errorf("function call %s %v not found", fc.Function.String(), err)
		}

	}()
	Logger.Println("Entered evalFunctionCall")
	function := i.Eval(fc.Function)
	args := []interface{}{}
	for _, arg := range fc.Arguments {
		argVal := i.Eval(arg)
		Logger.Println("Argument value:", argVal)
		args = append(args, argVal)
	}

	if fn, ok := function.(func([]interface{}) interface{}); ok {
		result := fn(args)
		Logger.Println("Function call result:", result)
		return result
	}
	Logger.Println("Function not callable:", function)
	return nil
}
