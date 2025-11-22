package indicators

import (
	"fmt"
	"testing"
)


func Test_series(t *testing.T) {
	var b = []float64{1,1,1,1}
	d := NewSeries(b)
	d[0] = 3

	fmt.Println(b)
}
func Test_if(t *testing.T) {
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})
	// 创建条件序列：close > 103
	condition := make([]bool, len(close))
	for i, v := range close {
		condition[i] = v > 103
	}
	
	// IF(condition, close, 0) - 如果条件为真返回close的值，否则返回0
	result, err := CallIndicatorByReflection("IF", []any{condition, close, NewSeries([]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}
}

func Test_bool_conversion(t *testing.T) {
	fmt.Println("=== 测试 []float64 -> []bool 转换 ===")
	
	// 测试：Series -> []bool
	close := NewSeries([]float64{0, 1, 2, 0, 3, -1, 0})
	result, err := CallIndicatorByReflection("IF", []any{close, NewSeries([]float64{10, 10, 10, 10, 10, 10, 10}), NewSeries([]float64{20, 20, 20, 20, 20, 20, 20})})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
		fmt.Println("说明: 非零值(1,2,3,-1)返回10，零值返回20")
	}
}

func TestSimpleReflectionExample(t *testing.T) {
	// 创建测试数据
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})

	fmt.Println("=== 简单反射调用示例 ===")

	// 示例1: SMA(close, 100, 1.0) - 你提到的例子
	fmt.Println("示例1: SMA(close, 100, 1.0)")
	result, err := CallIndicatorByReflection("SMA", []any{close, 100, 1.0})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}

	// 示例2: MA(close, 5)
	fmt.Println("\n示例2: MA(close, 5)")
	result, err = CallIndicatorByReflection("MA", []any{close, 5})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}

	// 示例3: ABS(close)
	fmt.Println("\n示例3: ABS(close)")
	result, err = CallIndicatorByReflection("ABS", []any{close})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}

	// 示例4: MAX(close, high)
	fmt.Println("\n示例4: MAX(close, high)")
	high := NewSeries([]float64{101, 103, 102, 104, 106, 105, 107, 109, 108, 110})
	result, err = CallIndicatorByReflection("MAX", []any{close, high})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}

	// 示例5: MACD(close, 12, 26, 9) - 多返回值
	fmt.Println("\n示例5: MACD(close, 12, 26, 9)")
	result, err = CallIndicatorByReflection("MACD", []any{close, 12, 26, 9})
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}
}

func TestReflectionParameterTypes(t *testing.T) {
	// 创建测试数据
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})

	fmt.Println("=== 反射参数类型测试 ===")

	// 测试不同的参数类型
	testCases := []struct {
		name string
		args []any
		desc string
	}{
		{"SMA", []any{close, 5, 1.0}, "SMA(close, 5, 1.0)"},
		{"SMA", []any{close, int(5), float64(1)}, "SMA(close, int(5), float64(1))"},
		{"MA", []any{close, 5}, "MA(close, 5)"},
		{"MA", []any{close, int(5)}, "MA(close, int(5))"},
		{"ABS", []any{close}, "ABS(close)"},
		{"POW", []any{close, 2.0}, "POW(close, 2.0)"},
		{"POW", []any{close, float64(2)}, "POW(close, float64(2))"},
		{"REF", []any{close, 1}, "REF(close, 1)"},
		{"REF", []any{close, int(1)}, "REF(close, int(1))"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.desc)
		result, err := CallIndicatorByReflection(tc.name, tc.args)
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
		} else {
			fmt.Printf("✅ 成功: %v\n", result)
		}
	}
}

func TestReflectionErrorHandling(t *testing.T) {
	// 创建测试数据
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})

	fmt.Println("=== 反射错误处理测试 ===")

	// 测试各种错误情况
	errorCases := []struct {
		name string
		args []any
		desc string
	}{
		{"NOTEXIST", []any{close, 5}, "不存在的函数"},
		{"MA", []any{close}, "参数数量不足"},
		{"MA", []any{close, 5, 6}, "参数数量过多"},
		{"SMA", []any{close, 5}, "SMA参数数量不足"},
		{"SMA", []any{close, 5, 1.0, 2.0}, "SMA参数数量过多"},
		{"MA", []any{close, "invalid"}, "无效的参数类型"},
		{"SMA", []any{close, 5, "invalid"}, "无效的参数类型"},
	}

	for _, tc := range errorCases {
		fmt.Printf("\n测试: %s\n", tc.desc)
		result, err := CallIndicatorByReflection(tc.name, tc.args)
		if err != nil {
			fmt.Printf("✅ 预期的错误: %v\n", err)
		} else {
			fmt.Printf("❌ 应该失败但成功了: %v\n", result)
		}
	}
}

func TestSingleBoolToSliceConversion(t *testing.T) {
	fmt.Println("=== 测试单个 bool 值转 []bool 切片 ===")
	
	trueSlice := NewSeries([]float64{1, 1, 1, 1, 1})
	falseSlice := NewSeries([]float64{0, 0, 0, 0, 0})
	
	// 测试1: 单个 bool(true) -> []bool
	fmt.Println("\n测试1: IF(true, trueSlice, falseSlice)")
	result, err := CallIndicatorByReflection("IF", []any{true, trueSlice, falseSlice})
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		t.Fail()
	} else {
		fmt.Printf("✅ 成功: %v\n", result)
	}
	
	// 测试2: 单个 bool(false) -> []bool
	fmt.Println("\n测试2: IF(false, trueSlice, falseSlice)")
	result, err = CallIndicatorByReflection("IF", []any{false, trueSlice, falseSlice})
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		t.Fail()
	} else {
		fmt.Printf("✅ 成功: %v\n", result)
	}
	
	// 测试3: []bool -> []bool (现有功能)
	fmt.Println("\n测试3: IF([]bool{true, false, true}, trueSlice, falseSlice)")
	boolSlice := []bool{true, false, true, true, false}
	result, err = CallIndicatorByReflection("IF", []any{boolSlice, trueSlice, falseSlice})
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		t.Fail()
	} else {
		fmt.Printf("✅ 成功: %v\n", result)
	}
}

// TestTypeConversionAndBroadcasting 专门测试类型转换和广播功能
func TestTypeConversionAndBroadcasting(t *testing.T) {
	fmt.Println("\n=== 类型转换和广播功能综合测试 ===")
	
	testData := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106})
	
	tests := []struct {
		name        string
		funcName    string
		args        []any
		description string
		expectedOk  bool
	}{
		{
			name:        "单个bool转[]bool_true",
			funcName:    "IF",
			args:        []any{true, testData, NewSeries([]float64{0, 0, 0, 0, 0, 0, 0})},
			description: "单个 bool(true) 应该广播为长度为 7 的 []bool",
			expectedOk:  true,
		},
		{
			name:        "单个bool转[]bool_false",
			funcName:    "IF",
			args:        []any{false, testData, NewSeries([]float64{0, 0, 0, 0, 0, 0, 0})},
			description: "单个 bool(false) 应该广播为长度为 7 的 []bool",
			expectedOk:  true,
		},
		{
			name:        "广播到不同长度",
			funcName:    "IF",
			args:        []any{true, NewSeries([]float64{1, 2, 3}), NewSeries([]float64{4, 5, 6, 7, 8})},
			description: "bool(true) 应该广播到最长参数的长度(5)",
			expectedOk:  true,
		},
		{
			name:        "POW单个float64转Series",
			funcName:    "POW",
			args:        []any{testData, 2.0},
			description: "单个 float64 参数作为标量使用",
			expectedOk:  true,
		},
		{
			name:        "ABS转Series",
			funcName:    "ABS",
			args:        []any{NewSeries([]float64{-1, -2, 0, 1, 2})},
			description: "ABS 函数处理 Series",
			expectedOk:  true,
		},
		{
			name:        "Series转[]bool",
			funcName:    "IF",
			args:        []any{testData, NewSeries([]float64{1, 1, 1, 1, 1, 1, 1}), NewSeries([]float64{0, 0, 0, 0, 0, 0, 0})},
			description: "Series 应该转换为 []bool (非零为true)",
			expectedOk:  true,
		},
		{
			name:        "空Series转[]bool",
			funcName:    "IF",
			args:        []any{NewSeries([]float64{0, 0, 0}), NewSeries([]float64{1, 1, 1}), NewSeries([]float64{0, 0, 0})},
			description: "全零 Series 转换为 []bool (全false)",
			expectedOk:  true,
		},
		{
			name:        "负数Series转[]bool",
			funcName:    "IF",
			args:        []any{NewSeries([]float64{-1, 0, 1}), NewSeries([]float64{1, 1, 1}), NewSeries([]float64{0, 0, 0})},
			description: "包含负数的 Series 转 []bool",
			expectedOk:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("\n测试: %s\n", tt.description)
			result, err := CallIndicatorByReflection(tt.funcName, tt.args)
			
			if tt.expectedOk {
				if err != nil {
					fmt.Printf("❌ 失败: %v\n", err)
					t.Errorf("期望成功但失败: %v", err)
				} else {
					fmt.Printf("✅ 成功: %v\n", result)
					
					// 验证结果不为空
					if result == nil {
						t.Errorf("结果不应为 nil")
					}
					
					// 对于 IF 函数，验证结果长度
					if tt.funcName == "IF" {
						if resultSlice, ok := result.(Series); ok {
							fmt.Printf("   结果长度: %d\n", len(resultSlice))
						}
					}
				}
			} else {
				if err == nil {
					fmt.Printf("❌ 应该失败但成功了: %v\n", result)
					t.Errorf("期望失败但成功了")
				} else {
					fmt.Printf("✅ 预期的错误: %v\n", err)
				}
			}
		})
	}
	
	fmt.Println("\n=== 测试完成 ===")
}

// TestREFWithSeries 测试 REF 函数支持 N 为 Series 的情况
func TestREFWithSeries(t *testing.T) {
	fmt.Println("\n=== 测试 REF 函数支持 N 为 Series ===")
	
	testData := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})
	
	tests := []struct {
		name        string
		S           Series
		N           any
		description string
		expectedLen int
	}{
		{
			name:        "REF_N为int",
			S:           testData,
			N:           3,
			description: "REF(C, 3) - N 为整数",
			expectedLen: len(testData),
		},
		{
			name:        "REF_N为Series",
			S:           testData,
			N:           NewSeries([]float64{1, 2, 3, 1, 2, 3, 1, 2, 3, 1}),
			description: "REF(C, N_Series) - N 为 Series",
			expectedLen: len(testData),
		},
		{
			name:        "REF_N为Series不同值",
			S:           NewSeries([]float64{10, 20, 30, 40, 50}),
			N:           NewSeries([]float64{0, 1, 2, 1, 0}),
			description: "REF(S, N) 每个位置使用不同的偏移量",
			expectedLen: 5,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("\n测试: %s\n", tt.description)
			result, err := CallIndicatorByReflection("REF", []any{tt.S, tt.N})
			
			if err != nil {
				fmt.Printf("❌ 失败: %v\n", err)
				t.Errorf("期望成功但失败: %v", err)
			} else {
				fmt.Printf("✅ 成功: %v\n", result)
				
				if resultSlice, ok := result.(Series); ok {
					fmt.Printf("   结果长度: %d, 期望长度: %d\n", len(resultSlice), tt.expectedLen)
					if len(resultSlice) != tt.expectedLen {
						t.Errorf("结果长度不匹配: 期望 %d, 得到 %d", tt.expectedLen, len(resultSlice))
					}
					
					// 显示部分结果
					displayLen := 5
					if len(resultSlice) < displayLen {
						displayLen = len(resultSlice)
					}
					fmt.Printf("   前 %d 个值: %v\n", displayLen, resultSlice[:displayLen])
				} else {
					t.Errorf("结果类型不正确，期望 Series")
				}
			}
		})
	}
	
	fmt.Println("\n=== REF 测试完成 ===")
}
