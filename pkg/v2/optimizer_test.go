package v2

import (
	"math"
	"testing"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/mylang"
)

func TestConstFold(t *testing.T) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("A:1+2*3;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	as := prog.Statements[0].(*mylang.AssignmentStatement)
	lit, ok := as.Value.(*mylang.NumberLiteral)
	if !ok {
		t.Fatalf("常量未折叠，类型=%T value=%s", as.Value, as.Value.String())
	}
	if lit.Value != 7 {
		t.Fatalf("1+2*3 = %v, want 7", lit.Value)
	}
}

func TestConstFoldDivByZero(t *testing.T) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("A:1/0;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	as := prog.Statements[0].(*mylang.AssignmentStatement)
	// 除零不可折叠，应保持 BinaryExpression
	if _, ok := as.Value.(*mylang.BinaryExpression); !ok {
		t.Fatalf("除零不应折叠，got %T", as.Value)
	}
}

func TestCSE(t *testing.T) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("A:=EMA(C,2); B:=EMA(C,2)+1; R:A+B;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	// CSE + DCE 后：__cse1 := EMA(C,2); A := __cse1; B := __cse1+1; R := A+B
	if len(prog.Statements) != 4 {
		t.Fatalf("语句数 = %d, want 4", len(prog.Statements))
	}
	first, ok := prog.Statements[0].(*mylang.AssignmentStatement)
	if !ok || first.Name.Value != "__cse1" || first.IsOutputVar {
		t.Fatalf("第 1 条应是中间变量 __cse1（非输出），got %+v", prog.Statements[0])
	}
	if fc, ok := first.Value.(*mylang.FunctionCall); !ok || fc.String() != "EMA(C, 2)" {
		t.Fatalf("__cse1 的值应是 EMA(C, 2)，got %v", first.Value.String())
	}
}

func TestSemanticEquiv(t *testing.T) {
	src := "A:=EMA(C,2); B:=EMA(C,2)+MA(C,3); R:A+B;"
	vars := map[string][]float64{
		"C": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	// 原版（不优化）
	ex1 := api.NewMaiExecutor()
	out1, err := ex1.Run(src, vars)
	if err != nil {
		t.Fatalf("原版 Run: %v", err)
	}

	// v2（优化）
	e2 := New()
	out2, err := e2.Run(src, vars)
	if err != nil {
		t.Fatalf("v2 Run: %v", err)
	}

	if len(out1) != len(out2) {
		t.Fatalf("输出数不一致: %d vs %d", len(out1), len(out2))
	}
	// 只有 R 是输出变量（A/B 是 := 中间变量）
	if out1[0].Name != "R" || out2[0].Name != "R" {
		t.Fatalf("输出名 = %q / %q, want R", out1[0].Name, out2[0].Name)
	}
	v1, v2 := out1[0].Values, out2[0].Values
	if len(v1) != len(v2) {
		t.Fatalf("长度不一致: %d vs %d", len(v1), len(v2))
	}
	for i := range v1 {
		a, b := v1[i], v2[i]
		if math.IsNaN(a) && math.IsNaN(b) {
			continue
		}
		if a != b {
			t.Fatalf("R[%d] = %v vs %v（语义不等价）", i, a, b)
		}
	}
}

func TestDCE(t *testing.T) {
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("A:=MA(C,5); B:=MA(C,10); R:B;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	// A 未被引用应被删：剩 B:=MA(C,10); R:B;
	if len(prog.Statements) != 2 {
		t.Fatalf("语句数 = %d, want 2", len(prog.Statements))
	}
	first := prog.Statements[0].(*mylang.AssignmentStatement)
	if first.Name.Value != "B" {
		t.Fatalf("第 1 条应是 B，got %s", first.Name.Value)
	}
}

func TestDCEKeepsReachable(t *testing.T) {
	// H1 被 H2 引用、H2 被 R 引用：中间链路必须全保留
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("H1:=MA(C,5); H2:=MA(H1,5); R:H2;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	if len(prog.Statements) != 3 {
		t.Fatalf("语句数 = %d, want 3（H1/H2 均可达）", len(prog.Statements))
	}
}

func TestDCESkipsReassign(t *testing.T) {
	// 同一变量重复赋值：DCE 保守跳过，保持 3 条原样
	ex := api.NewMaiExecutor()
	prog, err := ex.CompileSource("A:=MA(C,5); A:=MA(C,10); R:A;")
	if err != nil {
		t.Fatalf("CompileSource: %v", err)
	}
	Optimize(prog)
	if len(prog.Statements) != 3 {
		t.Fatalf("重复赋值应跳过 DCE，语句数 = %d, want 3", len(prog.Statements))
	}
}
