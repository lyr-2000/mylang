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
