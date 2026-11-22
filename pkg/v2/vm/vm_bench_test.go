package vm_test

import (
	"testing"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/v2/vm"
)

const benchSrc = "RSV:=(C-LLV(L,9))/(HHV(H,9)-LLV(L,9))*100; K:EMA(RSV,3); D:EMA(K,3); J:3*K-2*D;"

// BenchmarkOriginal 原包树遍历解释器。
func BenchmarkOriginal(t *testing.B) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource(benchSrc)
	if err != nil {
		t.Fatal(err)
	}
	vars := testVars(500)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := ex.RunCompiled(prog, vars); err != nil {
			t.Fatal(err)
		}
	}
}

// BenchmarkVM 字节码虚拟机。
func BenchmarkVM(t *testing.B) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource(benchSrc)
	if err != nil {
		t.Fatal(err)
	}
	chunk := vm.Compile(prog)
	v := buildVM(ex)
	vars := anyVars(testVars(500))
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := v.Run(chunk, vars); err != nil {
			t.Fatal(err)
		}
	}
}
