package vm_test

import (
	"math"
	"testing"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/v2/vm"
)

// buildVM 从原包执行器注入全部指标函数，使 vm 与原包共享同一份函数实现。
func buildVM(ex *api.MaiExecutor) *vm.VM {
	v := vm.NewVM()
	for _, name := range indicators.GetAllFuncNames() {
		if fn, ok := ex.Env.GetFunction(name); ok {
			if f, ok := fn.(func([]interface{}) interface{}); ok {
				v.SetFunc(name, f)
			}
		}
	}
	return v
}

func testVars(n int) map[string][]float64 {
	c := make([]float64, n)
	h := make([]float64, n)
	l := make([]float64, n)
	v := 100.0
	for i := range c {
		v += 0.3 + 0.01*float64(i%7)
		if i%5 == 0 {
			v -= 0.8
		}
		c[i] = v
		h[i] = v * 1.01
		l[i] = v * 0.99
	}
	return map[string][]float64{"C": c, "H": h, "L": l}
}

// TestSemanticEquiv 验证 vm 与原包对同一脚本的输出逐位一致。
func TestSemanticEquiv(t *testing.T) {
	srcs := []string{
		// KDJ 公开指标：算术 + 比较 + 函数调用 + 变量引用链
		"RSV:=(C-LLV(L,9))/(HHV(H,9)-LLV(L,9))*100; K:EMA(RSV,3); D:EMA(K,3); J:3*K-2*D;",
		// 标量算术 + 常量折叠不依赖
		"R:C+1-2*3/4;",
		// 数组 + 标量
		"R:(C-1)*100;",
		// 数组数组
		"R:MA(C,5)-MA(C,10);",
		// 比较（数组 + 数组 / 数组 + 标量）
		"A:MA(C,5)>MA(C,10); R:A*C;",
		"R:C>100;",
		// 逻辑
		"R:(C>100) AND (C<110);",
		"R:NOT(C>100);",
		// 函数调用
		"R:MAX(C,HHV(C,5));",
		"R:IF(C>100,C,0);",
		// 相等比较（floatEqual 语义）
		"R:C==C;",
	}
	vars := testVars(60)

	for _, src := range srcs {
		ex := api.NewMaiExecutor()
		outs1, err := ex.Run(src, vars)
		if err != nil {
			t.Fatalf("原包 %q: %v", src, err)
		}

		prog, err := ex.CompileSource(src)
		if err != nil {
			t.Fatalf("CompileSource %q: %v", src, err)
		}
		chunk := vm.Compile(prog)
		v := buildVM(ex)
		outs2, err := v.Run(chunk, anyVars(vars))
		if err != nil {
			t.Fatalf("vm %q: %v", src, err)
		}

		assertOutputsEq(t, src, outs1, outs2)
	}
}

func assertOutputsEq(t *testing.T, src string, a []api.Output, b []vm.Output) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("%q: 输出数 %d vs %d", src, len(a), len(b))
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			t.Fatalf("%q: 输出 %d 名 %s vs %s", src, i, a[i].Name, b[i].Name)
		}
		av := a[i].Values
		bv := vmValues(b[i])
		if len(av) != len(bv) {
			t.Fatalf("%q: 输出 %s 长度 %d vs %d", src, a[i].Name, len(av), len(bv))
		}
		for j := range av {
			if math.IsNaN(av[j]) && math.IsNaN(bv[j]) {
				continue
			}
			if av[j] != bv[j] {
				t.Fatalf("%q: 输出 %s[%d] = %v vs %v", src, a[i].Name, j, av[j], bv[j])
			}
		}
	}
}

func vmValues(o vm.Output) []float64 {
	switch v := o.Raw.(type) {
	case []float64:
		return v
	case indicators.Series:
		return []float64(v)
	case []interface{}: // 多返回值
		if len(v) > 0 {
			if first, ok := v[0].(indicators.Series); ok {
				return []float64(first)
			}
		}
	}
	return nil
}

func anyVars(vars map[string][]float64) map[string]interface{} {
	out := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		out[k] = v
	}
	return out
}
