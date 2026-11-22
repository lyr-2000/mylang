package v2

import (
	"testing"

	"github.com/lyr-2000/mylang/pkg/api"
)

// benchSrc 是「有冗余」的脚本：
//   - EMA(C,5) 出现两次（CSE 可消除一次）
//   - DEAD 未被 R 引用（DCE 可删除整条 MA(C,20)）
const benchSrc = `X:=EMA(C,5); Y:=EMA(C,5); DEAD:=MA(C,20); R:X+Y;`

func benchVars(n int) map[string][]float64 {
	c := make([]float64, n)
	for i := range c {
		c[i] = 10 + float64(i)*0.5
	}
	return map[string][]float64{"C": c}
}

// BenchmarkOriginal 原包：编译一次 + 复用 AST 执行（不优化）。
func BenchmarkOriginal(b *testing.B) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource(benchSrc)
	if err != nil {
		b.Fatal(err)
	}
	vars := benchVars(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ex.RunCompiled(prog, vars); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkV2  v2：编译 + 优化一次 + 复用优化后 AST 执行。
func BenchmarkV2(b *testing.B) {
	ex := New()
	prog, err := ex.Compile(benchSrc)
	if err != nil {
		b.Fatal(err)
	}
	vars := benchVars(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ex.RunCompiled(prog, vars); err != nil {
			b.Fatal(err)
		}
	}
}
