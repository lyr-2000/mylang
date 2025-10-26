package mylang

import (
	"reflect"
	"testing"
)

func TestMylangInterpreterSuffixParams(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})
	interp.RegisterVariable("LOW", []float64{95.0, 96.0, 97.0})

	tests := []struct {
		name           string
		code           string
		expectedParams map[string][]string
	}{
		{
			name: "单个变量带修饰符",
			code: "test1:HIGH>CLOSE,COLORRED;",
			expectedParams: map[string][]string{
				"test1": {"COLORRED"},
			},
		},
		{
			name: "多个变量带修饰符",
			code: `
test1:HIGH>CLOSE,COLORRED;
test2:HIGH<CLOSE,COLORBLUE,NODRAW;
test3:HIGH=CLOSE,COLORGREEN;
`,
			expectedParams: map[string][]string{
				"test1": {"COLORRED"},
				"test2": {"COLORBLUE", "NODRAW"},
				"test3": {"COLORGREEN"},
			},
		},
		{
			name: "混合赋值类型",
			code: `
drawing_var:HIGH>CLOSE,COLORRED;
normal_var:=HIGH+CLOSE,COLORBLUE;
`,
			expectedParams: map[string][]string{
				"drawing_var": {"COLORRED"},
				"normal_var":  {"COLORBLUE"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置解释器状态
			interp.Reset()

			// 重新注册变量
			interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
			interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})
			interp.RegisterVariable("LOW", []float64{95.0, 96.0, 97.0})

			// 执行代码
			interp.Execute(tt.code)

			// 检查修饰符
			allParams := interp.GetAllSuffixParams()
			if !reflect.DeepEqual(allParams, tt.expectedParams) {
				t.Errorf("GetAllSuffixParams() = %v, want %v", allParams, tt.expectedParams)
			}
		})
	}
}

func TestMylangInterpreterGetSuffixParams(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})

	code := `
test1:HIGH>CLOSE,COLORRED;
test2:HIGH<CLOSE,COLORBLUE,NODRAW;
test3:HIGH=CLOSE,COLORGREEN;
`

	// 执行代码
	interp.Execute(code)

	// 测试 GetSuffixParams
	tests := []struct {
		name     string
		varName  string
		expected []string
		exists   bool
	}{
		{
			name:     "存在的变量-单个修饰符",
			varName:  "test1",
			expected: []string{"COLORRED"},
			exists:   true,
		},
		{
			name:     "存在的变量-多个修饰符",
			varName:  "test2",
			expected: []string{"COLORBLUE", "NODRAW"},
			exists:   true,
		},
		{
			name:     "存在的变量-单个修饰符",
			varName:  "test3",
			expected: []string{"COLORGREEN"},
			exists:   true,
		},
		{
			name:     "不存在的变量",
			varName:  "nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "空字符串变量名",
			varName:  "",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, exists := interp.GetSuffixParams(tt.varName)
			if exists != tt.exists {
				t.Errorf("GetSuffixParams(%s) exists = %v, want %v", tt.varName, exists, tt.exists)
			}
			if !reflect.DeepEqual(params, tt.expected) {
				t.Errorf("GetSuffixParams(%s) = %v, want %v", tt.varName, params, tt.expected)
			}
		})
	}
}

func TestMylangInterpreterDrawingVariablesWithSuffixParams(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})

	code := `
drawing_var1:HIGH>CLOSE,COLORRED;
drawing_var2:HIGH<CLOSE,COLORBLUE,NODRAW;
normal_var:=HIGH+CLOSE,COLORGREEN;
`

	// 执行代码
	interp.Execute(code)

	// 检查画图变量
	drawingVars := interp.GetOutputVariableMap()
	expectedDrawingVars := map[string]struct{}{
		"drawing_var1": {},
		"drawing_var2": {},
	}

	if !reflect.DeepEqual(drawingVars, expectedDrawingVars) {
		t.Errorf("GetDrawingVariables() = %v, want %v", drawingVars, expectedDrawingVars)
	}

	// 检查画图变量标识
	if !interp.IsOutputVariable("drawing_var1") {
		t.Error("Expected drawing_var1 to be a drawing variable")
	}
	if !interp.IsOutputVariable("drawing_var2") {
		t.Error("Expected drawing_var2 to be a drawing variable")
	}
	if interp.IsOutputVariable("normal_var") {
		t.Error("Expected normal_var to NOT be a drawing variable")
	}

	// 检查修饰符
	allParams := interp.GetAllSuffixParams()
	expectedParams := map[string][]string{
		"drawing_var1": {"COLORRED"},
		"drawing_var2": {"COLORBLUE", "NODRAW"},
		"normal_var":   {"COLORGREEN"},
	}

	if !reflect.DeepEqual(allParams, expectedParams) {
		t.Errorf("GetAllSuffixParams() = %v, want %v", allParams, expectedParams)
	}
}

func TestMylangInterpreterReset(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})

	code := `
test1:HIGH>CLOSE,COLORRED;
test2:HIGH<CLOSE,COLORBLUE,NODRAW;
`

	// 执行代码
	interp.Execute(code)

	// 检查执行结果
	allParams := interp.GetAllSuffixParams()
	if len(allParams) == 0 {
		t.Error("Expected suffix params after execution")
	}

	drawingVars := interp.GetOutputVariableMap()
	if len(drawingVars) == 0 {
		t.Error("Expected drawing variables after execution")
	}

	// 重置解释器
	interp.Reset()

	// 检查重置后的状态
	allParamsAfterReset := interp.GetAllSuffixParams()
	if len(allParamsAfterReset) != 0 {
		t.Error("Expected empty suffix params after reset")
	}

	drawingVarsAfterReset := interp.GetOutputVariableMap()
	if len(drawingVarsAfterReset) != 0 {
		t.Error("Expected empty drawing variables after reset")
	}

	// 检查变量是否也被清除
	if _, exists := interp.GetVariable("HIGH"); exists {
		t.Error("Expected HIGH variable to be cleared after reset")
	}
	if _, exists := interp.GetVariable("CLOSE"); exists {
		t.Error("Expected CLOSE variable to be cleared after reset")
	}
}

func TestMylangInterpreterSyntaxValidation(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})

	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{
			name:        "有效语法-带分号",
			code:        "test:HIGH>CLOSE;",
			shouldError: false,
		},
		{
			name:        "有效语法-带修饰符和分号",
			code:        "test:HIGH>CLOSE,COLORRED;",
			shouldError: false,
		},
		{
			name:        "无效语法-缺少分号",
			code:        "test:HIGH>CLOSE",
			shouldError: true,
		},
		{
			name:        "无效语法-多个语句缺少分号",
			code:        "test1:HIGH>CLOSE\ntest2:HIGH<CLOSE;",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置解释器状态
			interp.Reset()

			// 重新注册变量
			interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
			interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})

			// 执行代码
			result := interp.Execute(tt.code)

			if tt.shouldError {
				// 对于应该出错的语法，期望有错误
				if !interp.HasSyntaxErrors() {
					t.Errorf("Expected syntax error for code: %s, but got none", tt.code)
				}
				if result != nil {
					t.Errorf("Expected nil result for syntax error, but got: %v", result)
				}
			} else {
				// 对于有效语法，期望能正常执行且无错误
				if interp.HasSyntaxErrors() {
					t.Errorf("Expected no syntax errors, but got: %v", interp.GetSyntaxErrors())
				}
				if result == nil {
					t.Errorf("Expected non-nil result for valid syntax")
				}
			}
		})
	}
}

func TestMylangInterpreterComplexExpressionsWithSuffixParams(t *testing.T) {
	interp := NewMylangInterpreter()

	// 注册测试变量和函数
	interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
	interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})
	interp.RegisterVariable("LOW", []float64{95.0, 96.0, 97.0})

	// 注册一个简单的函数
	interp.RegisterFunction("MA", func(args []interface{}) interface{} {
		if len(args) != 2 {
			return nil
		}
		data, ok1 := args[0].([]float64)
		n, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			return nil
		}
		// 简单的移动平均实现
		result := make([]float64, len(data))
		for i := range data {
			result[i] = data[i] * n
		}
		return result
	})

	tests := []struct {
		name           string
		code           string
		expectedParams map[string][]string
	}{
		{
			name: "复杂逻辑表达式",
			code: "result:HIGH>CLOSE AND LOW<CLOSE,COLORRED,NODRAW;",
			expectedParams: map[string][]string{
				"result": {"COLORRED", "NODRAW"},
			},
		},
		{
			name: "函数调用表达式",
			code: "ma_result:MA(CLOSE,5),COLORBLUE;",
			expectedParams: map[string][]string{
				"ma_result": {"COLORBLUE"},
			},
		},
		{
			name: "嵌套表达式",
			code: "complex:(HIGH+CLOSE)>(LOW*2),COLORGREEN,NODRAW;",
			expectedParams: map[string][]string{
				"complex": {"COLORGREEN", "NODRAW"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置解释器状态
			interp.Reset()

			// 重新注册变量和函数
			interp.RegisterVariable("HIGH", []float64{105.0, 106.0, 107.0})
			interp.RegisterVariable("CLOSE", []float64{100.0, 101.0, 102.0})
			interp.RegisterVariable("LOW", []float64{95.0, 96.0, 97.0})
			interp.RegisterFunction("MA", func(args []interface{}) interface{} {
				if len(args) != 2 {
					return nil
				}
				data, ok1 := args[0].([]float64)
				n, ok2 := args[1].(float64)
				if !ok1 || !ok2 {
					return nil
				}
				result := make([]float64, len(data))
				for i := range data {
					result[i] = data[i] * n
				}
				return result
			})

			// 执行代码
			interp.Execute(tt.code)

			// 检查修饰符
			allParams := interp.GetAllSuffixParams()
			if !reflect.DeepEqual(allParams, tt.expectedParams) {
				t.Errorf("GetAllSuffixParams() = %v, want %v", allParams, tt.expectedParams)
			}
		})
	}
}
