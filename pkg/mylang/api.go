package mylang

import "github.com/spf13/cast"

// MylangInterpreter 代表麦语言解释器的接口
type MylangInterpreter struct {
	Interp *Interpreter
	Env    *Environment
	Err    error
}

func (mi *MylangInterpreter) Reset() {
	mi.Env = NewEnvironment()
	mi.Interp = NewInterpreter(mi.Env)
	mi.Err = nil
}

// NewMylangInterpreter 创建一个新的麦语言解释器
func NewMylangInterpreter() *MylangInterpreter {
	env := NewEnvironment()
	interp := NewInterpreter(env)
	return &MylangInterpreter{Interp: interp, Env: env}
}

// RegisterVariable 注册一个变量到解释器
func (mi *MylangInterpreter) RegisterVariable(name string, value interface{}) {
	mi.Env.Set(name, value)
}

// RegisterFunction 注册一个用户自定义函数到解释器
func (mi *MylangInterpreter) RegisterFunction(name string, fn func([]interface{}) interface{}) {
	mi.Env.Set(name, fn)
}

// CompileCode 预编译麦语言代码，返回语法树
func (mi *MylangInterpreter) CompileCode(code string) *Program {
	lexer := NewLexer(code)
	parser := NewParser(lexer)
	return parser.ParseProgram()
}

// ExecuteProgram 执行语法树
func (mi *MylangInterpreter) ExecuteProgram(program *Program) interface{} {
	return mi.Interp.Eval(program)
}

// Execute 执行麦语言代码
func (mi *MylangInterpreter) Execute(code string) interface{} {
	lexer := NewLexer(code)
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	return mi.Interp.Eval(program)
}

// GetVariable 从环境中获取一个值用于调试
func (mi *MylangInterpreter) GetVariable(name string) (interface{}, bool) {
	return mi.Env.Get(name)
}
func (mi *MylangInterpreter) SetVar(name string, value interface{}) {
	mi.Env.Set(name, value)
}
func (mi *MylangInterpreter) GetVariableSlice(name string) ([]interface{}, bool) {
	b,ok:=  mi.Env.Get(name)
	if !ok {
		return nil, false
	}
	return cast.ToSlice(b), true
}
// GetDrawingVariables 获取所有画图变量
func (mi *MylangInterpreter) GetDrawingVariables() map[string]struct{} {
	return mi.Interp.drawingVars
}

// IsDrawingVariable 检查变量是否为画图变量
func (mi *MylangInterpreter) IsDrawingVariable(name string) bool {
	_, exists := mi.Interp.drawingVars[name]
	return exists
}
