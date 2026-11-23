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

// countingPool 记录 Get/Put 调用次数的 SeriesPool，用于验证 WithPool 注入生效。
type countingPool struct {
	gets int
	puts int
}

func (p *countingPool) Get(n int) []float64 { p.gets++; return make([]float64, n) }
func (p *countingPool) Put(s []float64)     { p.puts++ }

// TestWithPool 验证 WithPool 注入自定义 SeriesPool 后，列式运算走自定义池。
func TestWithPool(t *testing.T) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("R:(C-1)*100;")
	if err != nil {
		t.Fatal(err)
	}
	chunk := vm.Compile(prog)

	cp := &countingPool{}
	v := vm.NewVM(vm.WithPool(cp))
	if _, err := v.Run(chunk, map[string]interface{}{"C": make([]float64, 60)}); err != nil {
		t.Fatal(err)
	}
	if cp.gets == 0 {
		t.Fatal("WithPool 注入的自定义池未被使用（gets == 0）")
	}
	if cp.puts == 0 {
		t.Fatal("轮末未归还 buffer 到自定义池（puts == 0）")
	}
	t.Logf("自定义池 Get=%d Put=%d", cp.gets, cp.puts)
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

// TestMembufNoAliasing 验证 membuf 跨轮复用后，前一轮输出不被后一轮污染（无串味）。
func TestMembufNoAliasing(t *testing.T) {
	ex := api.NewMaiExecutor()
	src := "RSV:=(C-LLV(L,9))/(HHV(H,9)-LLV(L,9))*100; K:EMA(RSV,3); D:EMA(K,3); J:3*K-2*D;"
	prog, err := ex.CompileSource(src)
	if err != nil {
		t.Fatal(err)
	}
	chunk := vm.Compile(prog)
	v := buildVM(ex)

	vars1 := testVars(60)
	vars2 := testVars(60)
	for i := range vars2["C"] {
		vars2["C"][i] += 100 // 数据明显不同，串味会立刻暴露
	}

	out1, err := v.Run(chunk, anyVars(vars1))
	if err != nil {
		t.Fatal(err)
	}
	snap := make([][]float64, len(out1))
	for i, o := range out1 {
		snap[i] = append([]float64(nil), vmValues(o)...)
	}

	// 第二轮用不同数据跑，触发 membuf 复用第一轮归还的 buffer。
	if _, err := v.Run(chunk, anyVars(vars2)); err != nil {
		t.Fatal(err)
	}

	// 第一轮输出必须保持不变。
	for i, o := range out1 {
		got := vmValues(o)
		for j := range got {
			if math.IsNaN(snap[i][j]) && math.IsNaN(got[j]) {
				continue
			}
			if snap[i][j] != got[j] {
				t.Fatalf("串味：输出 %s[%d] 第一轮 %v → 被污染为 %v", o.Name, j, snap[i][j], got[j])
			}
		}
	}
}

func anyVars(vars map[string][]float64) map[string]interface{} {
	out := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		out[k] = v
	}
	return out
}
