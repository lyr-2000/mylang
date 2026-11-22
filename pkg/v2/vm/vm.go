package vm

import (
	"fmt"
	"math"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// Func 是注入的函数签名（与原包 RegisterFunction 的包装函数一致）。
// ContextFunc 由调用方包装成闭包后注入。
type Func func([]interface{}) interface{}

// Output 输出变量（Name + 原始值）。
type Output struct {
	Name string
	Raw  interface{}
}

// VM 字节码虚拟机。vars/funcs 通过 Run/SetFunc 注入，不依赖原包执行器。
type VM struct {
	funcs   map[string]Func
	vars    map[string]interface{}
	stack   []interface{}
	outputs []Output
	// VarGetter 可选：LOAD_VAR 查不到变量时回调（对应原包 CustomVariableGetter）。
	VarGetter func(name string) any
}

// NewVM 创建一个空虚拟机。
func NewVM() *VM {
	return &VM{
		funcs: map[string]Func{},
		vars:  map[string]interface{}{},
	}
}

// SetFunc 注入一个普通函数。
func (vm *VM) SetFunc(name string, fn Func) {
	vm.funcs[name] = fn
}

// GetVar 返回运行后保留在变量表里的值（中间变量/输出变量/注入变量），供调试与后续读取。
func (vm *VM) GetVar(name string) (any, bool) {
	v, ok := vm.vars[name]
	return v, ok
}

// Run 执行字节码，返回输出变量列表。vars 为注入的变量（如 CLOSE 序列）。
func (vm *VM) Run(chunk *Chunk, vars map[string]interface{}) ([]Output, error) {
	// 重置运行时状态，保留已注入的函数。
	vm.stack = vm.stack[:0]
	vm.outputs = vm.outputs[:0]
	for k := range vm.vars {
		delete(vm.vars, k)
	}
	for k, v := range vars {
		vm.vars[k] = v
	}

	for _, instr := range chunk.Code {
		switch instr.Op {
		case OpLoadConst:
			vm.push(chunk.Consts[instr.Idx])
		case OpLoadVar:
			name := chunk.Names[instr.Idx]
			val, ok := vm.vars[name]
			if !ok && vm.VarGetter != nil {
				if gv := vm.VarGetter(name); gv != nil {
					val, ok = gv, true
				}
			}
			if !ok {
				return nil, fmt.Errorf("变量未定义: %s", name)
			}
			vm.push(val)
		case OpStoreVar:
			vm.vars[chunk.Names[instr.Idx]] = vm.pop()
		case OpStoreOutput:
			name := chunk.Names[instr.Idx]
			val := vm.pop()
			vm.vars[name] = val
			vm.outputs = append(vm.outputs, Output{Name: name, Raw: val})
		case OpPop:
			vm.pop()
		case OpAdd, OpSub, OpMul, OpDiv, OpGT, OpLT, OpGE, OpLE, OpEQ, OpNE, OpAnd, OpOr:
			right := vm.pop()
			left := vm.pop()
			val, err := binOp(left, right, instr.Op)
			if err != nil {
				return nil, err
			}
			vm.push(val)
		case OpNot, OpNeg:
			vm.push(unaryOp(vm.pop(), instr.Op))
		case OpCall:
			name := chunk.Names[instr.Idx]
			fn, ok := vm.funcs[name]
			if !ok {
				return nil, fmt.Errorf("函数未定义: %s", name)
			}
			result := fn(vm.popN(instr.N))
			if err, ok := result.(error); ok {
				return nil, err
			}
			vm.push(result)
		case OpReturn:
			return vm.outputs, nil
		}
	}
	return vm.outputs, nil
}

// ---- 栈操作 ----

func (vm *VM) push(v interface{}) { vm.stack = append(vm.stack, v) }

func (vm *VM) pop() interface{} {
	n := len(vm.stack) - 1
	v := vm.stack[n]
	vm.stack = vm.stack[:n]
	return v
}

func (vm *VM) popN(n int) []interface{} {
	out := make([]interface{}, n)
	for i := n - 1; i >= 0; i-- {
		out[i] = vm.pop()
	}
	return out
}

// ---- 二元运算分派 ----

func isArith(op Opcode) bool   { return op >= OpAdd && op <= OpDiv }
func isCompare(op Opcode) bool { return op >= OpGT && op <= OpNE }

// binOp 复刻原包 evalBinaryExpression 的语义，但把 operator 的 switch 提到循环外。
func binOp(left, right interface{}, op Opcode) (interface{}, error) {
	if op == OpAnd || op == OpOr {
		return logicOp(left, right, op), nil
	}
	if isCompare(op) {
		return compareOp(left, right, op), nil
	}
	return arithOp(left, right, op)
}

// arithOp 算术：标量标量 / 数组标量 / 标量数组 / 数组数组（长度不匹配报错）。
func arithOp(left, right interface{}, op Opcode) (interface{}, error) {
	if l, ok := left.(float64); ok {
		if r, ok := right.(float64); ok {
			return scalarArith(l, r, op), nil
		}
	}
	larr, lok := asFloat64Slice(left)
	rarr, rok := asFloat64Slice(right)
	if lok && rok {
		if len(larr) != len(rarr) {
			return nil, fmt.Errorf("数组长度不匹配: %d != %d", len(larr), len(rarr))
		}
		return arrArrArith(larr, rarr, op), nil
	}
	if lok {
		if r, ok := right.(float64); ok {
			return arrScalarArith(larr, r, op), nil
		}
	}
	if rok {
		if l, ok := left.(float64); ok {
			return scalarArrArith(l, rarr, op), nil
		}
	}
	return nil, nil
}

func scalarArith(l, r float64, op Opcode) float64 {
	switch op {
	case OpAdd:
		return l + r
	case OpSub:
		return l - r
	case OpMul:
		return l * r
	case OpDiv:
		if r == 0 {
			return math.NaN()
		}
		return l / r
	}
	return 0
}

func arrArrArith(a, b []float64, op Opcode) []float64 {
	result := make([]float64, len(a))
	switch op {
	case OpAdd:
		for i := range a {
			result[i] = a[i] + b[i]
		}
	case OpSub:
		for i := range a {
			result[i] = a[i] - b[i]
		}
	case OpMul:
		for i := range a {
			result[i] = a[i] * b[i]
		}
	case OpDiv:
		for i := range a {
			if b[i] == 0 {
				result[i] = math.NaN()
			} else {
				result[i] = a[i] / b[i]
			}
		}
	}
	return result
}

func arrScalarArith(a []float64, s float64, op Opcode) []float64 {
	result := make([]float64, len(a))
	switch op {
	case OpAdd:
		for i := range a {
			result[i] = a[i] + s
		}
	case OpSub:
		for i := range a {
			result[i] = a[i] - s
		}
	case OpMul:
		for i := range a {
			result[i] = a[i] * s
		}
	case OpDiv:
		if s == 0 {
			for i := range a {
				result[i] = math.NaN()
			}
		} else {
			for i := range a {
				result[i] = a[i] / s
			}
		}
	}
	return result
}

func scalarArrArith(s float64, b []float64, op Opcode) []float64 {
	result := make([]float64, len(b))
	switch op {
	case OpAdd:
		for i := range b {
			result[i] = s + b[i]
		}
	case OpSub:
		for i := range b {
			result[i] = s - b[i]
		}
	case OpMul:
		for i := range b {
			result[i] = s * b[i]
		}
	case OpDiv:
		for i := range b {
			result[i] = s / b[i]
		}
	}
	return result
}

// compareOp 比较：标量返回 bool，数组返回 []float64（0/1）。
func compareOp(left, right interface{}, op Opcode) interface{} {
	larr, lok := asFloat64Slice(left)
	rarr, rok := asFloat64Slice(right)
	if lok && rok {
		return arrArrCompare(larr, rarr, op)
	}
	if lok {
		return arrScalarCompare(larr, toFloat64(right), op)
	}
	if rok {
		return arrScalarCompare(rarr, toFloat64(left), op)
	}
	return scalarCompare(toFloat64(left), toFloat64(right), op)
}

func scalarCompare(l, r float64, op Opcode) bool {
	switch op {
	case OpGT:
		return l > r
	case OpLT:
		return l < r
	case OpGE:
		return l >= r
	case OpLE:
		return l <= r
	case OpEQ:
		return floatEqual(l, r)
	case OpNE:
		return !floatEqual(l, r)
	}
	return false
}

func arrArrCompare(a, b []float64, op Opcode) []float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := make([]float64, n)
	switch op {
	case OpGT:
		for i := 0; i < n; i++ {
			if a[i] > b[i] {
				result[i] = 1
			}
		}
	case OpLT:
		for i := 0; i < n; i++ {
			if a[i] < b[i] {
				result[i] = 1
			}
		}
	case OpGE:
		for i := 0; i < n; i++ {
			if a[i] >= b[i] {
				result[i] = 1
			}
		}
	case OpLE:
		for i := 0; i < n; i++ {
			if a[i] <= b[i] {
				result[i] = 1
			}
		}
	case OpEQ:
		for i := 0; i < n; i++ {
			if floatEqual(a[i], b[i]) {
				result[i] = 1
			}
		}
	case OpNE:
		for i := 0; i < n; i++ {
			if !floatEqual(a[i], b[i]) {
				result[i] = 1
			}
		}
	}
	return result
}

func arrScalarCompare(a []float64, s float64, op Opcode) []float64 {
	result := make([]float64, len(a))
	switch op {
	case OpGT:
		for i := range a {
			if a[i] > s {
				result[i] = 1
			}
		}
	case OpLT:
		for i := range a {
			if a[i] < s {
				result[i] = 1
			}
		}
	case OpGE:
		for i := range a {
			if a[i] >= s {
				result[i] = 1
			}
		}
	case OpLE:
		for i := range a {
			if a[i] <= s {
				result[i] = 1
			}
		}
	case OpEQ:
		for i := range a {
			if floatEqual(a[i], s) {
				result[i] = 1
			}
		}
	case OpNE:
		for i := range a {
			if !floatEqual(a[i], s) {
				result[i] = 1
			}
		}
	}
	return result
}

// logicOp 逻辑 AND/OR：标量返回 bool，数组返回 []float64（0/1）。
func logicOp(left, right interface{}, op Opcode) interface{} {
	larr, lok := asFloat64Slice(left)
	rarr, rok := asFloat64Slice(right)
	if lok && rok {
		return arrArrLogic(larr, rarr, op)
	}
	if lok {
		return arrScalarLogic(larr, toBool(right), op)
	}
	if rok {
		return arrScalarLogic(rarr, toBool(left), op)
	}
	return scalarLogic(toBool(left), toBool(right), op)
}

func scalarLogic(l, r bool, op Opcode) bool {
	if op == OpAnd {
		return l && r
	}
	return l || r
}

func arrArrLogic(a, b []float64, op Opcode) []float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		lb, rb := a[i] != 0, b[i] != 0
		if (op == OpAnd && lb && rb) || (op == OpOr && (lb || rb)) {
			result[i] = 1
		}
	}
	return result
}

func arrScalarLogic(a []float64, s bool, op Opcode) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		b := a[i] != 0
		if (op == OpAnd && b && s) || (op == OpOr && (b || s)) {
			result[i] = 1
		}
	}
	return result
}

// unaryOp 一元运算：Not（!）与 Neg（-）。
func unaryOp(v interface{}, op Opcode) interface{} {
	if op == OpNeg {
		if s, ok := v.(float64); ok {
			return -s
		}
		if arr, ok := asFloat64Slice(v); ok {
			out := make([]float64, len(arr))
			for i := range arr {
				out[i] = -arr[i]
			}
			return out
		}
		return nil
	}
	// OpNot
	if arr, ok := asFloat64Slice(v); ok {
		out := make([]float64, len(arr))
		for i := range arr {
			if arr[i] == 0 {
				out[i] = 1
			}
		}
		return out
	}
	return !toBool(v)
}

// ---- 类型转换 ----

func asFloat64Slice(v interface{}) ([]float64, bool) {
	switch x := v.(type) {
	case []float64:
		return x, true
	case indicators.Series:
		return []float64(x), true
	}
	return nil, false
}

func toFloat64(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case uint8:
		return float64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case []float64:
		if len(x) > 0 {
			return x[0]
		}
		return 0
	case indicators.Series:
		if len(x) > 0 {
			return x[0]
		}
		return 0
	}
	return 0
}

func toBool(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case int:
		return x != 0
	case []float64:
		for _, e := range x {
			if e != 0 {
				return true
			}
		}
		return false
	case indicators.Series:
		for _, e := range x {
			if e != 0 {
				return true
			}
		}
		return false
	default:
		return x != nil
	}
}

// floatEqual 复刻原包的浮点近似相等（用于 == / !=）。
func floatEqual(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if math.Trunc(a) != math.Trunc(b) {
		return false
	}
	if math.Abs(a) < 1e-6 && math.Abs(b) < 1e-6 {
		return math.Abs(a-b) < 1e-6
	}
	if math.Abs(a-b) < 1e-10 {
		return true
	}
	maxVal := math.Max(math.Abs(a), math.Abs(b))
	if maxVal < 1e-10 {
		return true
	}
	return math.Abs(a-b)/maxVal < 0.01
}
