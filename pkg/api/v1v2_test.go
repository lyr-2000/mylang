package api

import (
	"math"
	"testing"
)

// TestV1V2Equiv 验证 v2（字节码 VM）与 v1（树遍历）对同一脚本输出逐位一致。
func TestV1V2Equiv(t *testing.T) {
	srcs := []string{
		// KDJ 公开指标：算术 + 函数调用 + 变量引用链
		kdjSrc,
		"R:MA(C,5)-MA(C,10);",
		"A:MA(C,5)>MA(C,10); R:A*C;",
		"R:(C>100) AND (C<110);",
		"R:IF(C>100,C,0);",
		"R:MAX(C,HHV(C,5));",
	}
	vars := kdjVars(60)

	for _, src := range srcs {
		outs2, err := NewMaiExecutor().Run(src, vars)
		if err != nil {
			t.Fatalf("v2 %q: %v", src, err)
		}
		outs1, err := NewMaiExecutorV1().Run(src, vars)
		if err != nil {
			t.Fatalf("v1 %q: %v", src, err)
		}
		if len(outs1) != len(outs2) {
			t.Fatalf("%q: 输出数 v1=%d v2=%d", src, len(outs1), len(outs2))
		}
		for i := range outs1 {
			if outs1[i].Name != outs2[i].Name {
				t.Fatalf("%q: 输出 %d 名 v1=%s v2=%s", src, i, outs1[i].Name, outs2[i].Name)
			}
			a, b := outs1[i].Values, outs2[i].Values
			if len(a) != len(b) {
				t.Fatalf("%q: 输出 %s 长度 v1=%d v2=%d", src, outs1[i].Name, len(a), len(b))
			}
			for j := range a {
				if math.IsNaN(a[j]) && math.IsNaN(b[j]) {
					continue
				}
				if a[j] != b[j] {
					t.Fatalf("%q: 输出 %s[%d] v1=%v v2=%v", src, outs1[i].Name, j, a[j], b[j])
				}
			}
		}
	}
}

func v1v2Seq(n int) []float64 {
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

// kdjSrc 是 KDJ 公开指标公式（RSV/K/D/J），用于测试与 benchmark，避免引入业务私有公式。
const kdjSrc = "RSV:=(C-LLV(L,9))/(HHV(H,9)-LLV(L,9))*100; K:EMA(RSV,3); D:EMA(K,3); J:3*K-2*D;"

// kdjVars 生成 KDJ 所需的 C/H/L 序列。
func kdjVars(n int) map[string][]float64 {
	c := v1v2Seq(n)
	h := make([]float64, n)
	l := make([]float64, n)
	for i := range c {
		h[i] = c[i] * 1.01
		l[i] = c[i] * 0.99
	}
	return map[string][]float64{"C": c, "H": h, "L": l}
}

// TestVMCustomVariableGetter 验证 vm 后端支持 CustomVariableGetter（原 v1 特有）。
func TestVMCustomVariableGetter(t *testing.T) {
	getter := func(name string) any {
		if name == "EXTRA" {
			return v1v2Seq(60)
		}
		return nil
	}

	outs2, err := NewMaiExecutor().runWithGetter(getter)
	if err != nil {
		t.Fatalf("v2: %v", err)
	}
	outs1, err := NewMaiExecutorV1().runWithGetter(getter)
	if err != nil {
		t.Fatalf("v1: %v", err)
	}
	assertSameOutputs(t, outs1, outs2)
}

func (m *MaiExecutor) runWithGetter(getter func(string) any) ([]Output, error) {
	m.SetCustomVariableGetter(getter)
	return m.Run("R:EXTRA*2;", nil)
}

// TestVMGetInternalVar 验证 vm 后端运行后可读中间变量（GetVariable/GetVariableSlice/GetFloat64Array）。
func TestVMGetInternalVar(t *testing.T) {
	vars := map[string][]float64{"C": v1v2Seq(60)}
	ex := NewMaiExecutor()
	if _, err := ex.Run("H1:=MA(C,5); R:H1;", vars); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, ok := ex.GetVariable("H1"); !ok {
		t.Fatal("vm 下 GetVariable 读不到中间变量 H1")
	}
	if v := ex.GetVariableSlice("H1"); len(v) != 60 {
		t.Fatalf("GetVariableSlice(H1) 长度 %d, want 60", len(v))
	}
	if v := ex.GetFloat64Array("H1"); len(v) != 60 {
		t.Fatalf("GetFloat64Array(H1) 长度 %d, want 60", len(v))
	}
}

func assertSameOutputs(t *testing.T, a, b []Output) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("输出数 %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Name != b[i].Name || len(a[i].Values) != len(b[i].Values) {
			t.Fatalf("输出 %d 不一致", i)
		}
		for j := range a[i].Values {
			if a[i].Values[j] != b[i].Values[j] {
				t.Fatalf("%s[%d] = %v vs %v", a[i].Name, j, a[i].Values[j], b[i].Values[j])
			}
		}
	}
}

const v1v2BenchSrc = kdjSrc

// BenchmarkV1 原树遍历解释器后端。
func BenchmarkV1(b *testing.B) {
	ex := NewMaiExecutorV1()
	prog, err := ex.CompileSource(v1v2BenchSrc)
	if err != nil {
		b.Fatal(err)
	}
	vars := kdjVars(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ex.RunCompiled(prog, vars); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkV2 字节码 VM 后端（默认 NewMaiExecutor）。
func BenchmarkV2(b *testing.B) {
	ex := NewMaiExecutor()
	prog, err := ex.CompileSource(v1v2BenchSrc)
	if err != nil {
		b.Fatal(err)
	}
	vars := kdjVars(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ex.RunCompiled(prog, vars); err != nil {
			b.Fatal(err)
		}
	}
}
