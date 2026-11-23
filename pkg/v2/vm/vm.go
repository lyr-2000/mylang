package vm

import (
	"fmt"
	"math"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/v2/membuf"
)

// Func 是注入的函数签名（与原包 RegisterFunction 的包装函数一致）。
// ContextFunc 由调用方包装成闭包后注入。
type Func func([]interface{}) interface{}

// Output 输出变量（Name + 原始值）。
type Output struct {
	Name string
	Raw  interface{}
}

// SeriesPool 列式运算中间 buffer 的池接口。面向接口编程：VM 只依赖该接口，
// 默认实现是 membuf.Pool，也可通过 WithPool 注入自定义实现。
type SeriesPool interface {
	Get(n int) []float64
	Put(s []float64)
}

// defaultSeriesPool 全局共享的默认 buffer 池。membuf.Pool 内部用 sync.Pool，
// 并发安全，多个 VM 共享同一池可提高 buffer 复用率、减少整体分配。
var defaultSeriesPool SeriesPool = membuf.New()

// Option 配置 VM 的可选项。
type Option func(*VM)

// WithPool 用自定义 SeriesPool 替换默认的全局共享池。
func WithPool(p SeriesPool) Option {
	return func(v *VM) {
		v.pool = p
	}
}

// VM 字节码虚拟机。vars/funcs 通过 Run/SetFunc 注入，不依赖原包执行器。
type VM struct {
	funcs   map[string]Func
	vars    map[string]interface{}
	stack   []interface{}
	outputs []Output
	// VarGetter 可选：LOAD_VAR 查不到变量时回调（对应原包 CustomVariableGetter）。
	VarGetter func(name string) any

	pool      SeriesPool     // 列式运算中间结果的 buffer 池
	allocated []pooledSeries // 本轮分配的中间 buffer（轮末统一归还）
}

// pooledSeries 标记「底层数组来自 membuf 池」的序列，与常规 []float64 区分，
// 便于 STORE 时逃逸复制、轮末归还。
type pooledSeries []float64

// NewVM 创建一个空虚拟机，默认使用全局共享的 membuf 池，可通过 WithPool 替换。
func NewVM(opts ...Option) *VM {
	v := &VM{
		funcs: map[string]Func{},
		vars:  map[string]interface{}{},
		pool:  defaultSeriesPool,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// SetPool 注入自定义 buffer 池（多个 VM 可共享同一池）。
func (vm *VM) SetPool(p SeriesPool) { vm.pool = p }

// allocSeries 从池分配长度为 n 的中间序列并记录，供轮末统一归还。
func (vm *VM) allocSeries(n int) pooledSeries {
	if vm.pool == nil {
		return pooledSeries(make([]float64, n))
	}
	s := pooledSeries(vm.pool.Get(n))
	vm.allocated = append(vm.allocated, s)
	return s
}

// escapeVal 逃逸复制：pooledSeries 复制到常规内存（作为变量/输出逃逸），其余原样返回。
func (vm *VM) escapeVal(v interface{}) interface{} {
	if ps, ok := v.(pooledSeries); ok {
		return append([]float64(nil), ps...)
	}
	return v
}

// release 归还本轮分配的所有中间 buffer。
func (vm *VM) release() {
	for _, s := range vm.allocated {
		if vm.pool != nil {
			vm.pool.Put(s)
		}
	}
	vm.allocated = vm.allocated[:0]
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
	vm.outputs = nil // 全新 slice：避免与上一轮返回的 outputs 共享底层数组（否则 append 会串味）
	vm.allocated = vm.allocated[:0]
	defer vm.release() // 轮末归还所有中间 buffer
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
				// 无参函数 fallback：裸标识符对应无参函数（如 HYBLOCK）。
				if fn, fok := vm.funcs[name]; fok {
					val = fn(nil)
					if _, isErr := val.(error); isErr {
						return nil, fmt.Errorf("变量未定义: %s", name)
					}
					ok = true
				}
			}
			if !ok {
				return nil, fmt.Errorf("变量未定义: %s", name)
			}
			vm.push(val)
		case OpStoreVar:
			vm.vars[chunk.Names[instr.Idx]] = vm.escapeVal(vm.pop())
		case OpStoreOutput:
			name := chunk.Names[instr.Idx]
			val := vm.escapeVal(vm.pop())
			vm.vars[name] = val
			vm.outputs = append(vm.outputs, Output{Name: name, Raw: val})
		case OpPop:
			vm.pop()
		case OpAdd, OpSub, OpMul, OpDiv, OpGT, OpLT, OpGE, OpLE, OpEQ, OpNE, OpAnd, OpOr:
			right := vm.pop()
			left := vm.pop()
			val, err := vm.binOp(left, right, instr.Op)
			if err != nil {
				return nil, err
			}
			vm.push(val)
		case OpNot, OpNeg:
			vm.push(vm.unaryOp(vm.pop(), instr.Op))
		case OpCall:
			name := chunk.Names[instr.Idx]
			fn, ok := vm.funcs[name]
			if !ok {
				return nil, fmt.Errorf("函数未定义: %s", name)
			}
			args := vm.popN(instr.N)
			for i, a := range args {
				if ps, ok := a.(pooledSeries); ok {
					args[i] = []float64(ps) // 零拷贝转常规切片，供注入函数（fastcall）识别
				}
			}
			result := fn(args)
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
func (vm *VM) binOp(left, right interface{}, op Opcode) (interface{}, error) {
	if op == OpAnd || op == OpOr {
		return vm.logicOp(left, right, op), nil
	}
	if isCompare(op) {
		return vm.compareOp(left, right, op), nil
	}
	return vm.arithOp(left, right, op)
}

// arithOp 算术：标量标量 / 数组标量 / 标量数组 / 数组数组（长度不匹配报错）。
func (vm *VM) arithOp(left, right interface{}, op Opcode) (interface{}, error) {
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
		return vm.arrArrArith(larr, rarr, op), nil
	}
	if lok {
		if r, ok := right.(float64); ok {
			return vm.arrScalarArith(larr, r, op), nil
		}
	}
	if rok {
		if l, ok := left.(float64); ok {
			return vm.scalarArrArith(l, rarr, op), nil
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

func (vm *VM) arrArrArith(a, b []float64, op Opcode) pooledSeries {
	result := vm.allocSeries(len(a))
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

func (vm *VM) arrScalarArith(a []float64, s float64, op Opcode) pooledSeries {
	result := vm.allocSeries(len(a))
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

func (vm *VM) scalarArrArith(s float64, b []float64, op Opcode) pooledSeries {
	result := vm.allocSeries(len(b))
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
func (vm *VM) compareOp(left, right interface{}, op Opcode) interface{} {
	larr, lok := asFloat64Slice(left)
	rarr, rok := asFloat64Slice(right)
	if lok && rok {
		return vm.arrArrCompare(larr, rarr, op)
	}
	if lok {
		return vm.arrScalarCompare(larr, toFloat64(right), op)
	}
	if rok {
		return vm.arrScalarCompare(rarr, toFloat64(left), op)
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

func (vm *VM) arrArrCompare(a, b []float64, op Opcode) pooledSeries {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := vm.allocSeries(n)
	clear(result) // 复用 buffer 可能残留旧值，0/1 结果需先清零
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

func (vm *VM) arrScalarCompare(a []float64, s float64, op Opcode) pooledSeries {
	result := vm.allocSeries(len(a))
	clear(result)
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
func (vm *VM) logicOp(left, right interface{}, op Opcode) interface{} {
	larr, lok := asFloat64Slice(left)
	rarr, rok := asFloat64Slice(right)
	if lok && rok {
		return vm.arrArrLogic(larr, rarr, op)
	}
	if lok {
		return vm.arrScalarLogic(larr, toBool(right), op)
	}
	if rok {
		return vm.arrScalarLogic(rarr, toBool(left), op)
	}
	return scalarLogic(toBool(left), toBool(right), op)
}

func scalarLogic(l, r bool, op Opcode) bool {
	if op == OpAnd {
		return l && r
	}
	return l || r
}

func (vm *VM) arrArrLogic(a, b []float64, op Opcode) pooledSeries {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := vm.allocSeries(n)
	clear(result)
	for i := 0; i < n; i++ {
		lb, rb := a[i] != 0, b[i] != 0
		if (op == OpAnd && lb && rb) || (op == OpOr && (lb || rb)) {
			result[i] = 1
		}
	}
	return result
}

func (vm *VM) arrScalarLogic(a []float64, s bool, op Opcode) pooledSeries {
	result := vm.allocSeries(len(a))
	clear(result)
	for i := range a {
		b := a[i] != 0
		if (op == OpAnd && b && s) || (op == OpOr && (b || s)) {
			result[i] = 1
		}
	}
	return result
}

// unaryOp 一元运算：Not（!）与 Neg（-）。
func (vm *VM) unaryOp(v interface{}, op Opcode) interface{} {
	if op == OpNeg {
		if s, ok := v.(float64); ok {
			return -s
		}
		if arr, ok := asFloat64Slice(v); ok {
			out := vm.allocSeries(len(arr))
			for i := range arr {
				out[i] = -arr[i]
			}
			return out
		}
		return nil
	}
	// OpNot
	if arr, ok := asFloat64Slice(v); ok {
		out := vm.allocSeries(len(arr))
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
	case pooledSeries:
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
	case pooledSeries:
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
	case pooledSeries:
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
