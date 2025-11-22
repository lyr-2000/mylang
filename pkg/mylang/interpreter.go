package mylang

import (
	"fmt"
	"io"
	"log"
	"math"
	stdlog "log"

	// "os"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/spf13/cast"
)

var (
	Logger = stdlog.New(io.Discard, "", stdlog.LstdFlags)
)

func SetLogger(logger *stdlog.Logger) {
	Logger = logger
}

// floatEqual 比较两个浮点数是否相等，允许 1% 的误差
// 但整数部分必须完全相等
func floatEqual(a, b float64) bool {
	// 如果两个数都是 NaN，则认为相等
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	// 如果一个是 NaN，另一个不是，则不相等
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	
	// 获取整数部分
	aInt := math.Trunc(a)
	bInt := math.Trunc(b)
	
	// 整数部分必须完全相等
	if aInt != bInt {
		return false
	}
	
	// 如果两个数都接近0，使用绝对误差
	if math.Abs(a) < 1e-6 && math.Abs(b) < 1e-6 {
		return math.Abs(a-b) < 1e-6
	}
	
	// 如果差异小于绝对误差阈值，认为相等
	if math.Abs(a-b) < 1e-10 {
		return true
	}
	
	// 使用相对误差，误差 < 1%（相对于整个数）
	maxVal := math.Max(math.Abs(a), math.Abs(b))
	if maxVal < 1e-10 {
		return true
	}
	
	relativeError := math.Abs(a-b) / maxVal
	return relativeError < 0.01
}

// Environment 代表执行环境
type Environment struct {
	variables map[string]interface{} // 存储变量
	functions map[string]interface{} // 存储函数
	outer     *Environment
}

func (e *Environment) DelAllVars() {
	x := e
	for x != nil {
		x.variables = make(map[string]interface{})
		x = x.outer
	}
}

// NewEnvironment 创建一个新的执行环境
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]interface{}),
		functions: make(map[string]interface{}),
	}
}

// Get 从环境中获取一个值（先在变量中查找，再在函数中查找）
func (e *Environment) Get(name string) (interface{}, bool) {
	// 先在变量中查找
	if val, ok := e.variables[name]; ok {
		return val, ok
	}
	// 再在函数中查找
	if val, ok := e.functions[name]; ok {
		return val, ok
	}
	// 如果都没找到且存在外层环境，则在外层环境中查找
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

// Set 在环境中设置一个值（默认设置为变量）
func (e *Environment) Set(name string, val interface{}) interface{} {
	e.variables[name] = val
	return val
}

// SetVariable 在环境中设置一个变量
func (e *Environment) SetVariable(name string, val interface{}) interface{} {
	e.variables[name] = val
	return val
}

// SetFunction 在环境中设置一个函数
func (e *Environment) SetFunction(name string, fn interface{}) interface{} {
	e.functions[name] = fn
	return fn
}

// GetVariable 从环境中获取一个变量
func (e *Environment) GetVariable(name string) (interface{}, bool) {
	if val, ok := e.variables[name]; ok {
		return val, ok
	}
	if e.outer != nil {
		return e.outer.GetVariable(name)
	}
	return nil, false
}

// GetFunction 从环境中获取一个函数
func (e *Environment) GetFunction(name string) (interface{}, bool) {
	if fn, ok := e.functions[name]; ok {
		return fn, ok
	}
	if e.outer != nil {
		return e.outer.GetFunction(name)
	}
	return nil, false
}

// Interpreter 代表解释器
type Interpreter struct {
	env                  *Environment
	CustomVariableGetter func(name string) any // if nil, use env.Get(name) to get the variable value
	OutputVarMap         map[string]int   // 记录画图变量
	_idx                 int                   //记录画图变量
	suffixParams         map[string][]string   // 记录变量名到修饰符的映射
	Err                  error
	SkipNilPointerCheck bool //if false, will panic on variable is nil
}
func (r *Interpreter) getOutputVariableId() int {
	r._idx++
	return r._idx
}

// NewInterpreter 创建一个新的解释器
func NewInterpreter(env *Environment) *Interpreter {
	return &Interpreter{
		env:          env,
		OutputVarMap: make(map[string]int),
		suffixParams: make(map[string][]string),
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
	case *UnaryExpression:
		Logger.Println("Evaluating unary expression")
		return i.evalUnaryExpression(node)
	case *NumberLiteral:
		Logger.Println("Evaluating number literal:", node.Value)
		return node.Value
	case *StringLiteral:
		Logger.Println("Evaluating string literal:", node.Value)
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
	if stmt.IsOutputVar {
		i.OutputVarMap[stmt.Name.Value] = i.getOutputVariableId()
		Logger.Println("Added", stmt.Name.Value, "to drawing variables")
	}

	// 存储修饰符
	if len(stmt.SuffixParams) > 0 {
		i.suffixParams[stmt.Name.Value] = stmt.SuffixParams
		Logger.Println("Added suffix params for", stmt.Name.Value, ":", stmt.SuffixParams)
	}

	return i.env.Set(stmt.Name.Value, val)
}

func (i *Interpreter) evalUnaryExpression(ue *UnaryExpression) interface{} {
	right := i.Eval(ue.Right)
	Logger.Println("Unary expression, operator:", ue.Operator, "right:", right)

	switch ue.Operator {
	case "NOT", "not":
		return i.evalLogicalNot(right)
	}

	return nil
}

func (i *Interpreter) evalBinaryExpression(be *BinaryExpression) interface{} {
	left := i.Eval(be.Left)
	right := i.Eval(be.Right)
	Logger.Println("Binary expression, left:", left, "operator:", be.Operator, "right:", right)

	// 处理逻辑运算
	switch be.Operator {
	case "AND":
		return i.evalLogicalAnd(left, right)
	case "OR", "or":
		return i.evalLogicalOr(left, right)
	}

	// 处理比较运算
	switch be.Operator {
	case ">":
		return i.evalComparison(left, right, ">")
	case "<":
		return i.evalComparison(left, right, "<")
	case ">=":
		return i.evalComparison(left, right, ">=")
	case "<=":
		return i.evalComparison(left, right, "<=")
	case "==", "=":
		return i.evalComparison(left, right, "==")
	case "!=":
		return i.evalComparison(left, right, "!=")
	}

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

	// 处理数组和浮点数的操作，兼容 Series 类型
	leftArr, larrOk := i.toFloat64Slice(left)
	rightArr, rarrOk := i.toFloat64Slice(right)

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
	if !i.SkipNilPointerCheck {
		log.Panicf("Variable Miss: %s",symbol.Value)
	}
	return nil
}

func (i *Interpreter) evalFunctionCall(fc *FunctionCall) interface{} {
	
	Logger.Println("Entered evalFunctionCall")
	function := i.Eval(fc.Function)
	args := []interface{}{}
	for _, arg := range fc.Arguments {
		argVal := i.Eval(arg)
		Logger.Println("Argument value:", argVal)
		args = append(args, argVal)
	}

	if fn, ok := function.(func([]interface{}) interface{}); ok {
		// defer func() {
		// 	err := recover()
		// 	if err != nil {
		// 		Logger.Printf("panic: %v", err)
		// 		i.Err = fmt.Errorf("function call %s %v not found", fc.Function.String(), err)
		// 	}
	
		// }()
		result := fn(args)
		Logger.Println("Function call result:", result)
		return result
	}
	if !i.SkipNilPointerCheck {
		log.Panic("Function not found ", fc.Function.String())
	}
	Logger.Println("Function not callable:", function)
	return nil
}

// evalLogicalAnd 实现逻辑 AND 运算
func (i *Interpreter) evalLogicalAnd(left, right interface{}) interface{} {
	// 处理数组模式，兼容 Series 类型
	leftArr, leftIsArr := i.toFloat64Slice(left)
	rightArr, rightIsArr := i.toFloat64Slice(right)

	if leftIsArr && rightIsArr {
		// 两个都是数组，逐元素进行 AND 运算
		return i.logicalAndArrays(leftArr, rightArr)
	} else if leftIsArr {
		// 左边是数组，右边是标量
		rightBool := i.toBool(right)
		return i.logicalAndArrayWithScalar(leftArr, rightBool)
	} else if rightIsArr {
		// 右边是数组，左边是标量
		leftBool := i.toBool(left)
		return i.logicalAndArrayWithScalar(rightArr, leftBool)
	}

	// 标量模式
	leftBool := i.toBool(left)
	rightBool := i.toBool(right)
	return leftBool && rightBool
}

// evalLogicalOr 实现逻辑 OR 运算
func (i *Interpreter) evalLogicalOr(left, right interface{}) interface{} {
	// 处理数组模式，兼容 Series 类型
	leftArr, leftIsArr := i.toFloat64Slice(left)
	rightArr, rightIsArr := i.toFloat64Slice(right)

	if leftIsArr && rightIsArr {
		// 两个都是数组，逐元素进行 OR 运算
		return i.logicalOrArrays(leftArr, rightArr)
	} else if leftIsArr {
		// 左边是数组，右边是标量
		rightBool := i.toBool(right)
		return i.logicalOrArrayWithScalar(leftArr, rightBool)
	} else if rightIsArr {
		// 右边是数组，左边是标量
		leftBool := i.toBool(left)
		return i.logicalOrArrayWithScalar(rightArr, leftBool)
	}

	// 标量模式
	leftBool := i.toBool(left)
	rightBool := i.toBool(right)
	return leftBool || rightBool
}

// toBool 将值转换为布尔值
func (i *Interpreter) toBool(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case int:
		return v != 0
	case string:
		return v != ""
	case []float64:
		// 对于数组，检查是否有任何非零元素
		for _, val := range v {
			if val != 0 {
				return true
			}
		}
		return false
	case indicators.Series:
		// 对于 Series，检查是否有任何非零元素
		for _, val := range v {
			if val != 0 {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// evalComparison 实现比较运算
func (i *Interpreter) evalComparison(left, right interface{}, operator string) interface{} {
	// 处理数组比较，兼容 Series 类型
	leftArr, leftIsArr := i.toFloat64Slice(left)
	rightArr, rightIsArr := i.toFloat64Slice(right)

	if leftIsArr && rightIsArr {
		// 两个都是数组，逐元素比较
		return i.compareArrays(leftArr, rightArr, operator)
	} else if leftIsArr {
		// 左边是数组，右边是标量
		rightVal := i.toFloat64(right)
		return i.compareArrayWithScalar(leftArr, rightVal, operator)
	} else if rightIsArr {
		// 右边是数组，左边是标量
		leftVal := i.toFloat64(left)
		return i.compareArrayWithScalar(rightArr, leftVal, operator)
	}

	// 标量比较
	leftVal := i.toFloat64(left)
	rightVal := i.toFloat64(right)

	switch operator {
	case ">":
		return leftVal > rightVal
	case "<":
		return leftVal < rightVal
	case ">=":
		return leftVal >= rightVal
	case "<=":
		return leftVal <= rightVal
	case "==", "=":
		return floatEqual(leftVal, rightVal)
	case "!=":
		return !floatEqual(leftVal, rightVal)
	default:
		return false
	}
}

// compareArrays 比较两个数组
func (i *Interpreter) compareArrays(leftArr, rightArr []float64, operator string) []float64 {
	minLen := len(leftArr)
	if len(rightArr) < minLen {
		minLen = len(rightArr)
	}

	result := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		switch operator {
		case ">":
			if leftArr[i] > rightArr[i] {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case "<":
			if leftArr[i] < rightArr[i] {
				result[i] = 1
			} else {
				result[i] = 0
			}
	case ">=":
		if leftArr[i] >= rightArr[i] {
			result[i] = 1
		} else {
			result[i] = 0
		}
	case "<=":
		if leftArr[i] <= rightArr[i] {
			result[i] = 1
		} else {
			result[i] = 0
		}
	case "==", "=":
		if floatEqual(leftArr[i], rightArr[i]) {
			result[i] = 1
		} else {
			result[i] = 0
		}
	case "!=":
		if !floatEqual(leftArr[i], rightArr[i]) {
			result[i] = 1
		} else {
			result[i] = 0
		}
		default:
			result[i] = 0
		}
	}
	return result
}

// compareArrayWithScalar 比较数组和标量
func (i *Interpreter) compareArrayWithScalar(arr []float64, scalar float64, operator string) []float64 {
	result := make([]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		switch operator {
		case ">":
			if arr[i] > scalar {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case "<":
			if arr[i] < scalar {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case ">=":
			if arr[i] >= scalar {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case "<=":
			if arr[i] <= scalar {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case "==", "=":
			if floatEqual(arr[i], scalar) {
				result[i] = 1
			} else {
				result[i] = 0
			}
		case "!=":
			if !floatEqual(arr[i], scalar) {
				result[i] = 1
			} else {
				result[i] = 0
			}
		default:
			result[i] = 0
		}
	}
	return result
}

// toFloat64Slice 将值转换为 []float64 切片，兼容 Series 类型
func (i *Interpreter) toFloat64Slice(value interface{}) ([]float64, bool) {
	if value == nil {
		return nil, false
	}

	switch v := value.(type) {
	case []float64:
		return v, true
	case indicators.Series:
		return []float64(v), true
	default:
		x := cast.ToFloat64Slice(value)
		return x, len(x) > 0
	}
}

// toFloat64 将值转换为浮点数
func (i *Interpreter) toFloat64(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case uint8:
		return float64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		return 0 // 字符串比较暂时返回0
	case []float64:
		if len(v) > 0 {
			return v[0] // 取数组第一个元素
		}
		return 0
	case indicators.Series:
		if len(v) > 0 {
			return v[0] // 取 Series 第一个元素
		}
		return 0
	default:
		return 0
	}
}

// logicalAndArrays 对两个数组进行逐元素 AND 运算
func (i *Interpreter) logicalAndArrays(leftArr, rightArr []float64) []float64 {
	minLen := len(leftArr)
	if len(rightArr) < minLen {
		minLen = len(rightArr)
	}

	result := make([]float64, minLen)
	for idx := 0; idx < minLen; idx++ {
		leftBool := leftArr[idx] != 0
		rightBool := rightArr[idx] != 0
		if leftBool && rightBool {
			result[idx] = 1
		} else {
			result[idx] = 0
		}
	}
	return result
}

// logicalAndArrayWithScalar 对数组和标量进行 AND 运算
func (i *Interpreter) logicalAndArrayWithScalar(arr []float64, scalar bool) []float64 {
	result := make([]float64, len(arr))
	for idx := 0; idx < len(arr); idx++ {
		arrBool := arr[idx] != 0
		if arrBool && scalar {
			result[idx] = 1
		} else {
			result[idx] = 0
		}
	}
	return result
}

// logicalOrArrays 对两个数组进行逐元素 OR 运算
func (i *Interpreter) logicalOrArrays(leftArr, rightArr []float64) []float64 {
	minLen := len(leftArr)
	if len(rightArr) < minLen {
		minLen = len(rightArr)
	}

	result := make([]float64, minLen)
	for idx := 0; idx < minLen; idx++ {
		leftBool := leftArr[idx] != 0
		rightBool := rightArr[idx] != 0
		if leftBool || rightBool {
			result[idx] = 1
		} else {
			result[idx] = 0
		}
	}
	return result
}

// logicalOrArrayWithScalar 对数组和标量进行 OR 运算
func (i *Interpreter) logicalOrArrayWithScalar(arr []float64, scalar bool) []float64 {
	result := make([]float64, len(arr))
	for idx := 0; idx < len(arr); idx++ {
		arrBool := arr[idx] != 0
		if arrBool || scalar {
			result[idx] = 1
		} else {
			result[idx] = 0
		}
	}
	return result
}

// evalLogicalNot 实现逻辑 NOT 运算
func (i *Interpreter) evalLogicalNot(right interface{}) interface{} {
	// 处理数组模式，兼容 Series 类型
	rightArr, rightIsArr := i.toFloat64Slice(right)

	if rightIsArr {
		// 数组，逐元素进行 NOT 运算
		return i.logicalNotArray(rightArr)
	}

	// 标量模式
	rightBool := i.toBool(right)
	if rightBool {
		return false
	}
	return true
}

// logicalNotArray 对数组进行逐元素 NOT 运算
func (i *Interpreter) logicalNotArray(arr []float64) []float64 {
	result := make([]float64, len(arr))
	for idx := 0; idx < len(arr); idx++ {
		arrBool := arr[idx] != 0
		if arrBool {
			result[idx] = 0
		} else {
			result[idx] = 1
		}
	}
	return result
}
