package api

import (
	"fmt"
	"github.com/lyr2000/mylang/pkg/extensions/indicators"
	"github.com/lyr2000/mylang/pkg/mylang"
	"io"
	"log"

	"github.com/spf13/cast"
)

type MaiExecutor struct {
	*mylang.MylangInterpreter
}

func (m *MaiExecutor) SetCustomVariableGetter(getter func (name string) (any)) {
	m.MylangInterpreter.Interp.CustomVariableGetter = getter
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

// func (m *MaiExecutor) RegisterFunc(name string, fn func([]interface{}) interface{}) {
// 	m.RegisterFunction(name, fn)
// }

// RunCode 执行麦语言代码，并返回执行结果字符串
func (m *MaiExecutor) RunCode(code string) error {
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
