package api

import (
	"fmt"
	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/mylang"
	"io"
	"log"
	"strings"

	"github.com/spf13/cast"
)

type MaiExecutor struct {
	*mylang.MylangInterpreter
	PreCompiledProgram *mylang.Program //if not nil ,use it to execute the program
	DateTimeKey        string
}

func (m *MaiExecutor) SetCustomVariableGetter(getter func(name string) any) {
	m.MylangInterpreter.Interp.CustomVariableGetter = getter
}

func (m *MaiExecutor) findOptionalDateTimeKey() string {
	var allKey = []string{
		"dateTime", "TS","Ts", "date", "Date",
	}
	for _, key := range allKey {
		_, ok := m.MylangInterpreter.GetVariable(key)
		if ok {
			return key
		}
	}
	return ""
}

func (m *MaiExecutor) GetDateTimeArray() []any {
	if m.DateTimeKey == "" {
		dateKey := m.findOptionalDateTimeKey()
		if dateKey == "" {
			dateKey = "dateTime"
		}
		m.DateTimeKey = dateKey
	}
	d, ok := m.MylangInterpreter.GetVariable(m.DateTimeKey)
	if !ok {
		return nil
	}
	d1, _ := mylang.ToSlice(d)
	return d1
}

func NewMaiExecutor() *MaiExecutor {
	d := &MaiExecutor{
		MylangInterpreter: mylang.NewMylangInterpreter(),
	}
	d.registerFuncs()
	return d
}

func SetOutput(output io.Writer) {
	mylang.SetLogger(log.New(output, "", log.LstdFlags))
}

func Arrayfloat64(b any) []float64 {
	return cast.ToFloat64Slice(b)
}

func (b *MaiExecutor) GetFloat64Array(name string) []float64 {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	return Arrayfloat64(d)
}

func (b *MaiExecutor) GetVariableSlice(name string) []any {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	d1, _ := mylang.ToSlice(d)
	return d1
}

// OPEN,HIGH,LOW,CLOSE,VOLUME,UnixSec -> O,H,L,C,V,Ts
func (b *MaiExecutor) SetVarNameAlias(alias map[string]string) {
	if alias == nil {
		alias = map[string]string{
			"OPEN":   "O",
			"HIGH":   "H",
			"LOW":    "L",
			"CLOSE":  "C",
			"VOLUME": "V",
		}
	}
	for name, alias := range alias {
		d, ok := b.MylangInterpreter.GetVariable(name)
		if ok {
			b.MylangInterpreter.SetVar(alias, d)
		}
	}

}

func (m *MaiExecutor) CompileCode(code string) error {
	m.PreCompiledProgram = m.MylangInterpreter.CompileCode(code)
	
	// 检查编译错误
	if len(m.PreCompiledProgram.Errors) > 0 {
		return fmt.Errorf("编译错误: %s", strings.Join(m.PreCompiledProgram.Errors, "; "))
	}
	
	return nil
}

func (m *MaiExecutor) PrintProgramTree() {
	if m.PreCompiledProgram == nil {
		fmt.Println("PreCompiledProgram is nil")
		return
	}
	
	fmt.Println("=== PreCompiledProgram AST ===")
	fmt.Printf("Total statements: %d\n", len(m.PreCompiledProgram.Statements))
	fmt.Println()
	
	for i, stmt := range m.PreCompiledProgram.Statements {
		fmt.Printf("Statement %d:\n", i)
		m.printStatement(stmt, 0)
		fmt.Println()
	}
	fmt.Println("=== End of AST ===")
}

func (m *MaiExecutor) printStatement(stmt mylang.Statement, indent int) {
	indentStr := strings.Repeat("  ", indent)
	
	switch s := stmt.(type) {
	case *mylang.AssignmentStatement:
		fmt.Printf("%sAssignmentStatement:\n", indentStr)
		fmt.Printf("%s  Name: %s\n", indentStr, s.Name.String())
		fmt.Printf("%s  IsDrawingVar: %t\n", indentStr, s.IsDrawingVar)
		fmt.Printf("%s  Value:\n", indentStr)
		m.printExpression(s.Value, indent+2)
		
	case *mylang.ExpressionStatement:
		fmt.Printf("%sExpressionStatement:\n", indentStr)
		fmt.Printf("%s  Expression:\n", indentStr)
		m.printExpression(s.Expression, indent+2)
		
	default:
		fmt.Printf("%sUnknown statement type: %T\n", indentStr, stmt)
		fmt.Printf("%s  String: %s\n", indentStr, stmt.String())
	}
}

func (m *MaiExecutor) printExpression(expr mylang.Expression, indent int) {
	if expr == nil {
		fmt.Printf("%s<nil>\n", strings.Repeat("  ", indent))
		return
	}
	
	indentStr := strings.Repeat("  ", indent)
	
	switch e := expr.(type) {
	case *mylang.Identifier:
		fmt.Printf("%sIdentifier: %s\n", indentStr, e.Value)
		
	case *mylang.NumberLiteral:
		fmt.Printf("%sNumberLiteral: %f\n", indentStr, e.Value)
		
	case *mylang.StringLiteral:
		fmt.Printf("%sStringLiteral: %s\n", indentStr, e.Value)
		
	case *mylang.BinaryExpression:
		fmt.Printf("%sBinaryExpression:\n", indentStr)
		fmt.Printf("%s  Operator: %s\n", indentStr, e.Operator)
		fmt.Printf("%s  Left:\n", indentStr)
		m.printExpression(e.Left, indent+2)
		fmt.Printf("%s  Right:\n", indentStr)
		m.printExpression(e.Right, indent+2)
		
	case *mylang.FunctionCall:
		fmt.Printf("%sFunctionCall:\n", indentStr)
		fmt.Printf("%s  Function:\n", indentStr)
		m.printExpression(e.Function, indent+2)
		fmt.Printf("%s  Arguments (%d):\n", indentStr, len(e.Arguments))
		for i, arg := range e.Arguments {
			fmt.Printf("%s    [%d]:\n", indentStr, i)
			m.printExpression(arg, indent+3)
		}
		
	default:
		fmt.Printf("%sUnknown expression type: %T\n", indentStr, expr)
		fmt.Printf("%s  String: %s\n", indentStr, expr.String())
	}
}

func (m *MaiExecutor) ExecuteProgram() error {
	if m.PreCompiledProgram == nil {
		return fmt.Errorf("PreCompiledProgram is nil")
	}
	if len(m.PreCompiledProgram.Errors) > 0 {
		log.Panicf("编译错误: %s", strings.Join(m.PreCompiledProgram.Errors, "; "))
	}
	m.MylangInterpreter.ExecuteProgram(m.PreCompiledProgram)
	return m.Err
}

// RunCode 执行麦语言代码，并返回执行结果字符串
func (m *MaiExecutor) RunCode(code string) error {
	if m.PreCompiledProgram != nil {
		log.Panicf("MaiExecutor is already compiled ,use ExecuteProgram() to execute the program")
	}
	// 假设这里可以调用核心麦语言解释器，实际应用中你需要替换为正确调用
	// 例如: result, err := mytt.RunMaiCode(code)
	m.Execute(code)
	// mylang.Logger.Printf("Result: %v", result)
	return m.Err
}

func callBasicFunc(name string, args []any) (ret any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	b, err := indicators.CallIndicatorByReflection(name, args)
	return b, err
}

func (m *MaiExecutor) registerFuncs() {
	for _, name := range indicators.GetAllFuncNames() {
		m.RegisterFunction(name, func(args []interface{}) interface{} {
			b, err := callBasicFunc(name, args)
			if err != nil {
				return err
			}
			return b
		})
	}

}
