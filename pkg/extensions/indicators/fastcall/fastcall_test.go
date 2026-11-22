package fastcall

import (
	"math"
	"testing"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// TestInvokerCoverage 验证 fastFuncs 里所有函数签名都命中类型化调用器。
func TestInvokerCoverage(t *testing.T) {
	for name := range fastFuncs {
		if _, ok := Invoker(name); !ok {
			t.Errorf("Invoker(%q) = false, want true（签名未命中）", name)
		}
	}
}

// TestInvokerSemanticEquiv 验证去反射调用与反射调用（CallIndicatorByReflection）结果逐位一致。
func TestInvokerSemanticEquiv(t *testing.T) {
	s := seq(20)
	cases := []struct {
		name string
		args []any
	}{
		{"EMA", []any{s, 12}},
		{"MA", []any{s, 5}},
		{"REF", []any{s, 5}},
		{"BARSLASTCOUNT", []any{[]bool{true, false, true, true, false}}},
		{"CROSS", []any{s, shift(s, 1)}},
		{"IF", []any{[]bool{true, false, true, false, true}, s, shift(s, 1)}},
		// 标量广播：IF 的第三参数为标量，应广播到 Series
		{"IF", []any{[]bool{true, false, true, false, true}, s, 0.0}},
		{"IF", []any{[]bool{true, false, true, false, true}, s, 5}},
		{"CCI", []any{s, shift(s, 1), shift(s, 2), 14}},
	}
	for _, c := range cases {
		inv, ok := Invoker(c.name)
		if !ok {
			t.Fatalf("%s 未覆盖", c.name)
		}
		got, err := inv(c.args)
		if err != nil {
			t.Fatalf("%s invoker: %v", c.name, err)
		}
		want, err := indicators.CallIndicatorByReflection(c.name, c.args)
		if err != nil {
			t.Fatalf("%s reflect: %v", c.name, err)
		}
		assertEqual(t, c.name, got, want)
	}
}

// assertEqual 比较返回值（Series / []bool），NaN 视为相等。
func assertEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	switch w := want.(type) {
	case indicators.Series:
		g, ok := got.(indicators.Series)
		if !ok {
			t.Fatalf("%s: got 类型 %T, want Series", name, got)
		}
		if len(g) != len(w) {
			t.Fatalf("%s: 长度 %d vs %d", name, len(g), len(w))
		}
		for i := range w {
			if math.IsNaN(w[i]) && math.IsNaN(g[i]) {
				continue
			}
			if w[i] != g[i] {
				t.Fatalf("%s[%d] = %v, want %v", name, i, g[i], w[i])
			}
		}
	case []bool:
		g, ok := got.([]bool)
		if !ok {
			t.Fatalf("%s: got 类型 %T, want []bool", name, got)
		}
		if len(g) != len(w) {
			t.Fatalf("%s: 长度 %d vs %d", name, len(g), len(w))
		}
		for i := range w {
			if g[i] != w[i] {
				t.Fatalf("%s[%d] = %v, want %v", name, i, g[i], w[i])
			}
		}
	default:
		if got != want {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
	}
}

func seq(n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = 10 + float64(i)*0.5
	}
	return s
}

func shift(s []float64, k int) []float64 {
	out := make([]float64, len(s))
	for i := range s {
		if i >= k {
			out[i] = s[i-k]
		}
	}
	return out
}

// TestRegister 验证用户自定义函数可注册并去反射调用。
func TestRegister(t *testing.T) {
	wrap, ok := Register("MySum", func(s indicators.Series, n int) indicators.Series {
		out := make(indicators.Series, len(s))
		for i, v := range s {
			out[i] = v + float64(n)
		}
		return out
	})
	if !ok {
		t.Fatal("Register 失败")
	}
	got := wrap([]any{seq(5), 10})
	s, ok := got.(indicators.Series)
	if !ok {
		t.Fatalf("返回类型 %T, want Series", got)
	}
	base := seq(5)
	for i, v := range s {
		if v != base[i]+10 {
			t.Fatalf("MySum[%d] = %v, want %v", i, v, base[i]+10)
		}
	}
}

// TestRegisterRejectOverride 验证注册时拒绝覆盖已有（系统）函数。
func TestRegisterRejectOverride(t *testing.T) {
	if _, ok := Register("EMA", indicators.EMA); ok {
		t.Fatal("覆盖系统函数 EMA 应失败")
	}
}

// TestRegisterUnsupportedSignature 验证不支持的签名返回 ok=false。
func TestRegisterUnsupportedSignature(t *testing.T) {
	if _, ok := Register("MyStr", func(a string) string { return a }); ok {
		t.Fatal("不支持签名应返回 ok=false")
	}
}
