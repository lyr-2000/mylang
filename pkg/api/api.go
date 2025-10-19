package api

import (
	"fmt"
	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/mylang"
	"io"
	"log"

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
	return nil
}

func (m *MaiExecutor) ExecuteProgram() error {
	if m.PreCompiledProgram == nil {
		return fmt.Errorf("PreCompiledProgram is nil")
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
