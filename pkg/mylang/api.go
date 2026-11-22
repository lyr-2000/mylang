package mylang

import (
	"reflect"
	"sort"
	"strings"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// MylangInterpreter 代表麦语言解释器的接口
type MylangInterpreter struct {
	Interp *Interpreter
	Env    *Environment
	Err    error
}

func (mi *MylangInterpreter) DelVars() {
	mi.Env.DelAllVars()
}

// ResetRuntimeState 只清理一次执行产生的状态（画图变量、修饰符、错误），
// 保留已注册的变量、函数与 CustomVariableGetter。
// 适合宿主复用预编译 Program 时，在每次 ExecuteProgram 前调用。
func (mi *MylangInterpreter) ResetRuntimeState() {
	mi.Interp.OutputVarMap = make(map[string]int)
	mi.Interp.suffixParams = make(map[string][]string)
	mi.Interp._idx = 0
	mi.Interp.Err = nil
	mi.Err = nil
}

// Reset 清空变量与执行状态，但保留函数和解释器配置
// （CustomVariableGetter / Context / SkipNilPointerCheck）。
// 函数表 functions 初始化后只读，无需像旧实现那样拷贝重建 Environment 和 Interpreter。
func (mi *MylangInterpreter) Reset() {
	mi.DelVars()
	mi.ResetRuntimeState()
}

// SetContext 设置本次执行上下文，横截面函数依赖它解析当前品种。
func (mi *MylangInterpreter) SetContext(ctx Context) {
	mi.Interp.Context = ctx
}

// RegisterContextFunction 注册可读取执行上下文的宿主函数。
func (mi *MylangInterpreter) RegisterContextFunction(name string, fn ContextFunc) {
	mi.Env.SetFunction(name, fn)
}

// NewMylangInterpreter 创建一个新的麦语言解释器
func NewMylangInterpreter() *MylangInterpreter {
	env := NewEnvironment()
	interp := NewInterpreter(env)
	return &MylangInterpreter{Interp: interp, Env: env}
}

// RegisterVariable 注册一个变量到解释器
func (mi *MylangInterpreter) RegisterVariable(name string, value interface{}) {
	mi.Env.SetVariable(name, value)
}

// RegisterFunction 注册一个用户自定义函数到解释器
func (mi *MylangInterpreter) RegisterFunction(name string, fn func([]interface{}) interface{}) {
	mi.Env.SetFunction(name, fn)
}

func hasComment(code string) bool {
	return strings.Contains(code, "{") || strings.Contains(code, "//")
}

// CompileError 携带解析器给出的全部诊断信息，便于宿主定位脚本问题。
type CompileError struct {
	Errors []string
}

func (e *CompileError) Error() string {
	return "语法错误: " + strings.Join(e.Errors, "; ")
}

// Compile 预编译脚本并返回可复用的语法树；编译失败时返回 *CompileError。
// 与 CompileCode 不同，本方法不修改执行器状态，调用方可以自行缓存 program。
func (mi *MylangInterpreter) Compile(code string) (*Program, error) {
	program := mi.CompileCode(code)
	if len(program.Errors) > 0 {
		return nil, &CompileError{Errors: program.Errors}
	}
	return program, nil
}

// CompileCode 预编译麦语言代码，返回语法树
func (mi *MylangInterpreter) CompileCode(code string) *Program {
	if hasComment(code) {
		code = TrimComment(code)
	}
	lexer := NewLexer(code)
	parser := NewParser(lexer)
	return parser.ParseProgram()
}

// ExecuteProgram 执行语法树
func (mi *MylangInterpreter) ExecuteProgram(program *Program) interface{} {
	result := mi.Interp.Eval(program)
	if mi.Interp.Err != nil {
		mi.Err = mi.Interp.Err
	}
	return result
}

// Execute 执行麦语言代码
func (mi *MylangInterpreter) Execute(code string) interface{} {
	program, err := mi.Compile(code)
	if err != nil {
		mi.Interp.Err = err
		mi.Err = err
		return nil
	}
	return mi.ExecuteProgram(program)
}

// GetVariable 从环境中获取一个值用于调试
func (mi *MylangInterpreter) GetVariable(name string) (interface{}, bool) {
	return mi.Env.Get(name)
}
func (mi *MylangInterpreter) SetVar(name string, value interface{}) {
	mi.Env.SetVariable(name, value)
}

func (mi *MylangInterpreter) GetVariableSlice(name string) ([]interface{}, bool) {
	b, ok := mi.Env.Get(name)
	if !ok {
		return nil, false
	}
	return ToSlice(b)
}

func ToSlice(b any) ([]any, bool) {
	sl, ok := b.([]string)
	if ok {
		return copySlice(sl), true
	}
	slf, ok := b.([]float64)
	if ok {
		return copySlice(slf), true
	}
	slf2, ok := b.(indicators.Series)
	if ok {
		return copySlice[float64](slf2), true
	}
	return AnySlice(b)
}

func AnySlice(b any) ([]any, bool) {
	switch b := b.(type) {
	case []string:
		return copySlice(b), true
	case []float64:
		return copySlice(b), true
	case indicators.Series:
		return copySlice[float64](b), true
	default:
		bVal := reflect.ValueOf(b)
		if bVal.Kind() != reflect.Slice {
			return nil, false
		}
		length := bVal.Len()
		var arr []any
		for i := 0; i < length; i++ {
			arr = append(arr, bVal.Index(i).Interface())
		}
		return arr, true
	}
}

func copySlice[T any](arr []T) []any {
	do := make([]any, len(arr))
	for i, v := range arr {
		do[i] = v
	}
	return do
}

// GetOutputVariableMap 获取所有画图变量
// map[string]int, string为变量名， int为输出的顺序， 要获取变量值需要用 GetOutputVariable(int) 获取
func (mi *MylangInterpreter) GetOutputVariableMap() map[string]int {
	return mi.Interp.OutputVarMap
}

// OutputVariable 是画图变量的元信息；Order 从 1 开始，与脚本中 ":" 赋值顺序一致。
type OutputVariable struct {
	Name         string
	Order        int
	SuffixParams []string
}

// OrderedOutputVariables 按脚本声明顺序返回画图变量。
// 调用方不应依赖 map 遍历顺序来还原指标序列。
func (mi *MylangInterpreter) OrderedOutputVariables() []OutputVariable {
	vars := make([]OutputVariable, 0, len(mi.Interp.OutputVarMap))
	for name, order := range mi.Interp.OutputVarMap {
		params, _ := mi.Interp.suffixParams[name]
		vars = append(vars, OutputVariable{
			Name:         name,
			Order:        order,
			SuffixParams: append([]string(nil), params...),
		})
	}
	sort.Slice(vars, func(a, b int) bool { return vars[a].Order < vars[b].Order })
	return vars
}

// GetOutputVariable 获取第 i 个输出值
func (mi *MylangInterpreter) GetOutputVariable(i int) (any, bool) {
	for name, id2 := range mi.GetOutputVariableMap() {
		if id2 == i {
			return mi.GetVariable(name)
		}
	}
	return nil, false
}

// GetLastOutput 获取最后一个输出
func (mi *MylangInterpreter) GetLastOutput() (any, bool) {
	last := 0
	// key := ""
	km := mi.GetOutputVariableMap()
	for _, v := range km {
		if v > last {
			last = v
			// key = k
		}
	}
	return mi.GetOutputVariable(last)
}

// IsOutputVariable 检查变量是否为画图变量
func (mi *MylangInterpreter) IsOutputVariable(name string) bool {
	_, exists := mi.Interp.OutputVarMap[name]
	return exists
}

// GetAllSuffixParams 获取所有变量的修饰符映射
func (mi *MylangInterpreter) GetAllSuffixParams() map[string][]string {
	return mi.Interp.suffixParams
}

// GetSuffixParams 获取指定变量的修饰符
func (mi *MylangInterpreter) GetSuffixParams(name string) ([]string, bool) {
	params, exists := mi.Interp.suffixParams[name]
	return params, exists
}

// HasSyntaxErrors 检查是否有语法错误
func (mi *MylangInterpreter) HasSyntaxErrors() bool {
	return mi.Interp.Err != nil
}

// GetSyntaxErrors 获取语法错误列表
func (mi *MylangInterpreter) GetSyntaxErrors() []string {
	if mi.Interp.Err != nil {
		return []string{mi.Interp.Err.Error()}
	}
	return []string{}
}

// GetAllVariables 获取所有变量
func (mi *MylangInterpreter) GetAllVariables() map[string]interface{} {
	return mi.Env.variables
}

// GetAllFunctions 获取所有函数
func (mi *MylangInterpreter) GetAllFunctions() map[string]interface{} {
	return mi.Env.functions
}

// GetVariableOnly 只从变量中获取值（不查找函数）
func (mi *MylangInterpreter) GetVariableOnly(name string) (interface{}, bool) {
	return mi.Env.GetVariable(name)
}

// GetFunctionOnly 只从函数中获取值（不查找变量）
func (mi *MylangInterpreter) GetFunctionOnly(name string) (interface{}, bool) {
	return mi.Env.GetFunction(name)
}
