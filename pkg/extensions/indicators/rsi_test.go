package indicators

import (
	"fmt"
	"testing"
)

// TestRSI 测试 RSI 指标的计算
func TestRSI(t *testing.T) {
	// 创建测试数据
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109, 110, 112, 111, 113, 115})
	
	fmt.Println("=== 测试 RSI 指标 ===")
	
	// 步骤1: LC := REF(CLOSE, 1)
	fmt.Println("\n步骤1: LC := REF(CLOSE, 1)")
	lc, err := CallIndicatorByReflection("REF", []any{close, 1})
	if err != nil {
		t.Fatalf("REF 调用失败: %v", err)
	}
	fmt.Printf("LC: %v\n", lc)
	
	// 步骤2: CLOSE - LC
	fmt.Println("\n步骤2: CLOSE - LC")
	closeMinusLC := make(Series, len(close))
	lcSeries := lc.(Series)
	for i := range close {
		if i < len(lcSeries) {
			closeMinusLC[i] = close[i] - lcSeries[i]
		} else {
			closeMinusLC[i] = 0
		}
	}
	fmt.Printf("CLOSE - LC: %v\n", closeMinusLC)
	
	// 步骤3: MAX(CLOSE-LC, 0) - 这里测试标量转换
	fmt.Println("\n步骤3: MAX(CLOSE-LC, 0)")
	maxResult, err := CallIndicatorByReflection("MAX", []any{closeMinusLC, 0.0})
	if err != nil {
		t.Fatalf("MAX 调用失败: %v", err)
	}
	fmt.Printf("MAX(CLOSE-LC, 0): %v\n", maxResult)
	
	// 步骤4: ABS(CLOSE-LC)
	fmt.Println("\n步骤4: ABS(CLOSE-LC)")
	absResult, err := CallIndicatorByReflection("ABS", []any{closeMinusLC})
	if err != nil {
		t.Fatalf("ABS 调用失败: %v", err)
	}
	fmt.Printf("ABS(CLOSE-LC): %v\n", absResult)
	
	// 步骤5: SMA(MAX(CLOSE-LC,0), N1, 1) 其中 N1=14
	fmt.Println("\n步骤5: SMA(MAX(CLOSE-LC,0), 14, 1)")
	sma1, err := CallIndicatorByReflection("SMA", []any{maxResult, 14, 1.0})
	if err != nil {
		t.Fatalf("SMA 调用失败: %v", err)
	}
	fmt.Printf("SMA(MAX(CLOSE-LC,0), 14, 1): %v\n", sma1)
	
	// 步骤6: SMA(ABS(CLOSE-LC), N1, 1)
	fmt.Println("\n步骤6: SMA(ABS(CLOSE-LC), 14, 1)")
	sma2, err := CallIndicatorByReflection("SMA", []any{absResult, 14, 1.0})
	if err != nil {
		t.Fatalf("SMA 调用失败: %v", err)
	}
	fmt.Printf("SMA(ABS(CLOSE-LC), 14, 1): %v\n", sma2)
	
	// 步骤7: RSI1 = SMA1 / SMA2 * 100
	fmt.Println("\n步骤7: RSI1 = SMA1 / SMA2 * 100")
	sma1Series := sma1.(Series)
	sma2Series := sma2.(Series)
	rsi1 := make(Series, len(sma1Series))
	for i := range sma1Series {
		if i < len(sma2Series) && sma2Series[i] != 0 {
			rsi1[i] = sma1Series[i] / sma2Series[i] * 100
		} else {
			rsi1[i] = 0
		}
	}
	fmt.Printf("RSI1: %v\n", rsi1)
}

// TestMAXWithScalar 测试 MAX 函数与标量的交互
func TestMAXWithScalar(t *testing.T) {
	fmt.Println("=== 测试 MAX 函数与标量 ===")
	
	close := NewSeries([]float64{100, 102, 101, 103, 105})
	
	// 测试 MAX(close, 102) - 应该使用 MAXS
	fmt.Println("\n测试: MAX(close, 102.0)")
	result, err := CallIndicatorByReflection("MAX", []any{close, 102.0})
	if err != nil {
		fmt.Printf("MAX 调用失败（预期，因为 MAX 需要两个 Series）: %v\n", err)
		// 尝试使用 MAXS
		fmt.Println("\n尝试使用 MAXS(close, 102.0)")
		result, err = CallIndicatorByReflection("MAXS", []any{close, 102.0})
		if err != nil {
			t.Fatalf("MAXS 调用失败: %v", err)
		}
		fmt.Printf("MAXS(close, 102.0): %v\n", result)
	} else {
		fmt.Printf("MAX(close, 102.0): %v\n", result)
	}
}

