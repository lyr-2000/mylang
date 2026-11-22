package api

import (
	"math"
	"testing"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// TestUserDefinedMACDLiteral 验证：用户用「字面量参数」手写 MACD（EMA 组合），
// 结果与系统 MACD 一致，且输出变量名 MACD 不与系统函数名 MACD 冲突。
func TestUserDefinedMACDLiteral(t *testing.T) {
	closes := macdSeq(60)
	ex := NewMaiExecutor()
	userSrc := "DIF:EMA(CLOSE,12)-EMA(CLOSE,26); DEA:EMA(DIF,9); MACD:(DIF-DEA)*2,COLORSTICK;"
	userOuts, err := ex.Run(userSrc, map[string][]float64{"CLOSE": closes})
	if err != nil {
		t.Fatalf("用户 MACD 脚本失败: %v", err)
	}
	userByName := map[string][]float64{}
	for _, o := range userOuts {
		userByName[o.Name] = o.Values
	}
	if len(userByName) != 3 {
		t.Fatalf("输出数 = %d, want 3", len(userByName))
	}

	ex2 := NewMaiExecutor()
	sysOuts, err := ex2.Run("R:MACD(CLOSE,12,26,9);", map[string][]float64{"CLOSE": closes})
	if err != nil {
		t.Fatalf("系统 MACD 脚本失败: %v", err)
	}
	parts, ok := sysOuts[0].Raw.([]any)
	if !ok || len(parts) != 3 {
		t.Fatalf("系统 MACD 多返回值结构异常: %T", sysOuts[0].Raw)
	}

	assertSeriesEq(t, "DIF", userByName["DIF"], []float64(parts[0].(indicators.Series)))
	assertSeriesEq(t, "DEA", userByName["DEA"], []float64(parts[1].(indicators.Series)))
	assertSeriesEq(t, "MACD", userByName["MACD"], []float64(parts[2].(indicators.Series)))
}

// TestUserDefinedMACDVarParam 验证：用「变量参数」（SHORT 等）时，
// 参数经 vars 传入是 []float64，而 EMA 要 int，[]float64→int 转换不被支持。
func TestUserDefinedMACDVarParam(t *testing.T) {
	ex := NewMaiExecutor()
	userSrc := "DIF:EMA(CLOSE,SHORT)-EMA(CLOSE,LONG); DEA:EMA(DIF,MID); MACD:(DIF-DEA)*2,COLORSTICK;"
	userVars := map[string][]float64{
		"CLOSE": macdSeq(30),
		"SHORT": {12},
		"LONG":  {26},
		"MID":   {9},
	}
	_, err := ex.Run(userSrc, userVars)
	if err == nil {
		t.Fatalf("预期变量参数版本失败（[]float64 无法转 int），但成功了")
	}
	t.Logf("变量参数版报错（符合预期）: %v", err)

	// 反射版直接验证：确认 []float64→int 在原 convertType 里同样不支持（非去反射引入）。
	if _, rerr := indicators.CallIndicatorByReflection("EMA", []any{macdSeq(30), []float64{12}}); rerr == nil {
		t.Fatal("反射版预期同样失败（[]float64 无法转 int），但成功了")
	} else {
		t.Logf("反射版同样报错（确认非去反射引入）: %v", rerr)
	}
}

func assertSeriesEq(t *testing.T, name string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s 长度 %d vs %d", name, len(got), len(want))
	}
	for i := range want {
		if math.IsNaN(got[i]) && math.IsNaN(want[i]) {
			continue
		}
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %v, want %v", name, i, got[i], want[i])
		}
	}
}

func macdSeq(n int) []float64 {
	s := make([]float64, n)
	v := 100.0
	for i := range s {
		v += 0.3 + 0.01*float64(i%7)
		if i%5 == 0 {
			v -= 0.8
		}
		s[i] = v
	}
	return s
}
