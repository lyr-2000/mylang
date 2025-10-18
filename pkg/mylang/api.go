package mylang

// MylangInterpreter 代表麦语言解释器的接口
type MylangInterpreter struct {
	interp *Interpreter
	env    *Environment
	Err    error
}

// NewMylangInterpreter 创建一个新的麦语言解释器
func NewMylangInterpreter() *MylangInterpreter {
	env := NewEnvironment()
	interp := NewInterpreter(env)
	return &MylangInterpreter{interp: interp, env: env}
}

// RegisterVariable 注册一个变量到解释器
func (mi *MylangInterpreter) RegisterVariable(name string, value interface{}) {
	mi.env.Set(name, value)
}

// RegisterFunction 注册一个用户自定义函数到解释器
func (mi *MylangInterpreter) RegisterFunction(name string, fn func([]interface{}) interface{}) {
	mi.env.Set(name, fn)
}

// Execute 执行麦语言代码
func (mi *MylangInterpreter) Execute(code string) interface{} {
	lexer := NewLexer(code)
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	return mi.interp.Eval(program)
}

// GetVariable 从环境中获取一个值用于调试
func (mi *MylangInterpreter) GetVariable(name string) (interface{}, bool) {
	return mi.env.Get(name)
}
func (mi *MylangInterpreter) SetVar(name string, value interface{}) {
	mi.env.Set(name, value)
}

// GetDrawingVariables 获取所有画图变量
func (mi *MylangInterpreter) GetDrawingVariables() map[string]struct{} {
	return mi.interp.drawingVars
}

// IsDrawingVariable 检查变量是否为画图变量
func (mi *MylangInterpreter) IsDrawingVariable(name string) bool {
	_, exists := mi.interp.drawingVars[name]
	return exists
}
