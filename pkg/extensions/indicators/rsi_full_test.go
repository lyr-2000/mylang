package indicators

import (
	"fmt"
	"testing"
)

// TestRSIFull 完整测试 RSI 指标的执行流程
func TestRSIFull(t *testing.T) {
	fmt.Println("=== 完整测试 RSI 指标执行流程 ===")
	
	// 创建测试数据
	close := NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109, 110, 112, 111, 113, 115})
	low := NewSeries([]float64{99, 101, 100, 102, 104, 103, 105, 107, 106, 108, 109, 111, 110, 112, 114})
	
	// 模拟 RSI 计算
	// LC := REF(CLOSE, 1)
	lc, err := CallIndicatorByReflection("REF", []any{close, 1})
	if err != nil {
		t.Fatalf("REF 调用失败: %v", err)
	}
	lcSeries := lc.(Series)
	fmt.Printf("LC: %v\n", lcSeries)
	
	// CLOSE - LC
	closeMinusLC := make(Series, len(close))
	for i := range close {
		if i < len(lcSeries) {
			closeMinusLC[i] = close[i] - lcSeries[i]
		} else {
			closeMinusLC[i] = 0
		}
	}
	fmt.Printf("CLOSE - LC: %v\n", closeMinusLC)
	
	// MAX(CLOSE-LC, 0)
	maxResult, err := CallIndicatorByReflection("MAX", []any{closeMinusLC, 0.0})
	if err != nil {
		t.Fatalf("MAX 调用失败: %v", err)
	}
	maxSeries := maxResult.(Series)
	fmt.Printf("MAX(CLOSE-LC, 0): %v\n", maxSeries)
	
	// ABS(CLOSE-LC)
	absResult, err := CallIndicatorByReflection("ABS", []any{closeMinusLC})
	if err != nil {
		t.Fatalf("ABS 调用失败: %v", err)
	}
	absSeries := absResult.(Series)
	fmt.Printf("ABS(CLOSE-LC): %v\n", absSeries)
	
	// SMA(MAX(CLOSE-LC,0), 14, 1)
	sma1, err := CallIndicatorByReflection("SMA", []any{maxSeries, 14, 1.0})
	if err != nil {
		t.Fatalf("SMA 调用失败: %v", err)
	}
	sma1Series := sma1.(Series)
	fmt.Printf("SMA(MAX(CLOSE-LC,0), 14, 1): %v\n", sma1Series)
	
	// SMA(ABS(CLOSE-LC), 14, 1)
	sma2, err := CallIndicatorByReflection("SMA", []any{absSeries, 14, 1.0})
	if err != nil {
		t.Fatalf("SMA 调用失败: %v", err)
	}
	sma2Series := sma2.(Series)
	fmt.Printf("SMA(ABS(CLOSE-LC), 14, 1): %v\n", sma2Series)
	
	// RSI1 = SMA1 / SMA2 * 100
	rsi1 := make(Series, len(sma1Series))
	for i := range sma1Series {
		if i < len(sma2Series) && sma2Series[i] != 0 {
			rsi1[i] = sma1Series[i] / sma2Series[i] * 100
		} else {
			rsi1[i] = 0
		}
	}
	fmt.Printf("RSI1: %v\n", rsi1)
	
	// RSI1 > 50 - 这应该返回 []float64
	fmt.Println("\n测试: RSI1 > 50")
	// 模拟比较操作：RSI1 > 50
	condition := make([]float64, len(rsi1))
	for i, v := range rsi1 {
		if v > 50 {
			condition[i] = 1.0
		} else {
			condition[i] = 0.0
		}
	}
	fmt.Printf("RSI1 > 50: %v\n", condition)
	
	// L * 0.9
	fmt.Println("\n测试: L * 0.9")
	lowTimes09 := make(Series, len(low))
	for i, v := range low {
		lowTimes09[i] = v * 0.9
	}
	fmt.Printf("L * 0.9: %v\n", lowTimes09)
	
	// BUY(RSI1>50, L*0.9)
	fmt.Println("\n测试: BUY(RSI1>50, L*0.9)")
	// 注意：这里需要将 condition 和 lowTimes09 转换为 []float64
	conditionFloat64 := make([]float64, len(condition))
	copy(conditionFloat64, condition)
	lowTimes09Float64 := make([]float64, len(lowTimes09))
	copy(lowTimes09Float64, lowTimes09)
	
	buyResult, err := CallIndicatorByReflection("BUY", []any{conditionFloat64, lowTimes09Float64})
	if err != nil {
		t.Fatalf("BUY 调用失败: %v", err)
	}
	buySeries := buyResult.([]float64)
	fmt.Printf("BUY(RSI1>50, L*0.9): %v\n", buySeries)
}

