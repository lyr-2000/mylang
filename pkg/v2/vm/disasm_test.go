package vm_test

import (
	"testing"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/v2/vm"
)

// TestDisasm 打印各脚本的「源代码 + 编译后的字节码」，便于查看编译产物。
func TestDisasm(t *testing.T) {
	srcs := []string{
		// KDJ 公开指标（算术 + 比较 + 函数调用 + 变量引用链）
		"RSV:=(C-LLV(L,9))/(HHV(H,9)-LLV(L,9))*100; K:EMA(RSV,3); D:EMA(K,3); J:3*K-2*D;",
		// 标量算术
		"R:C+1-2*3/4;",
		// 数组 + 标量
		"R:(C-1)*100;",
		// 数组数组
		"R:MA(C,5)-MA(C,10);",
		// 比较（数组 + 数组）
		"A:MA(C,5)>MA(C,10); R:A*C;",
		// 逻辑
		"R:(C>100) AND (C<110);",
		"R:NOT(C>100);",
		// 函数调用 + 标量广播
		"R:IF(C>100,C,0);",
	}

	ex := api.NewMaiExecutor()
	for _, src := range srcs {
		prog, err := ex.CompileSource(src)
		if err != nil {
			t.Fatalf("CompileSource %q: %v", src, err)
		}
		chunk := vm.Compile(prog)
		t.Logf("\n=== 源代码 ===\n%s\n=== 字节码 ===\n%s", src, chunk.Disassemble())
	}
}
