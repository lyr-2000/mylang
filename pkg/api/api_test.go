package api

import (
	"math"
	"testing"
)

func TestMaiExecutorCompileError(t *testing.T) {
	executor := NewMaiExecutor()

	// 注册测试变量
	executor.SetVar("C", []float64{100.0, 101.0, 102.0})
	executor.SetVar("HIGH", []float64{105.0, 106.0, 107.0})

	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{
			name:        "有效语法-带分号",
			code:        "test:HIGH>C;",
			shouldError: false,
		},
		{
			name:        "有效语法-带修饰符和分号",
			code:        "test:HIGH>C,COLORRED;",
			shouldError: false,
		},
		{
			name:        "无效语法-缺少分号",
			code:        "test:HIGH>C",
			shouldError: true,
		},
		{
			name:        "无效语法-多个语句缺少分号",
			code:        "test1:HIGH>C\ntest2:HIGH<C;",
			shouldError: true,
		},
		{
			name:        "charttest问题代码",
			code:        "去除:=1\nZTPPrice:ZTPRICE(C,0.1)\n涨停:=C>=ZTPRICE(REF(C,1),0.1);",
			shouldError: true,
		},
		{
			name:        "多行代码错误-验证行号",
			code:        "line1:HIGH>C;\nline2:HIGH<C\nline3:HIGH=C;",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置执行器状态
			executor = NewMaiExecutor()
			executor.SetVar("C", []float64{100.0, 101.0, 102.0})
			executor.SetVar("HIGH", []float64{105.0, 106.0, 107.0})

			// 编译代码
			err := executor.CompileCode(tt.code)

			if tt.shouldError {
				// 对于应该出错的语法，期望编译期报错
				if err == nil {
					t.Errorf("Expected compile error for code: %s, but got none", tt.code)
				}
			} else {
				// 对于有效语法，期望编译成功
				if err != nil {
					t.Errorf("Expected no compile error, but got: %v", err)
				}
			}
		})
	}
}

func TestMaiExecutorCompileSuccess(t *testing.T) {
	executor := NewMaiExecutor()

	// 注册测试变量
	executor.SetVar("C", []float64{100.0, 101.0, 102.0})
	executor.SetVar("HIGH", []float64{105.0, 106.0, 107.0})

	// 测试有效的代码
	code := `
test1:HIGH>C;
test2:HIGH<C,COLORRED;
test3:=C+1;
`

	// 编译代码
	err := executor.CompileCode(code)
	if err != nil {
		t.Fatalf("Expected no compile error, but got: %v", err)
	}

	// 检查预编译程序
	if executor.PreCompiledProgram == nil {
		t.Error("Expected PreCompiledProgram to be set after successful compilation")
	}

	if len(executor.PreCompiledProgram.Statements) == 0 {
		t.Error("Expected statements to be parsed")
	}

	// 检查是否有错误
	if len(executor.PreCompiledProgram.Errors) > 0 {
		t.Errorf("Expected no errors in compiled program, but got: %v", executor.PreCompiledProgram.Errors)
	}
}

func testBars() map[string][]float64 {
	return map[string][]float64{
		"OPEN":   []float64{1, 2, 3},
		"HIGH":   []float64{2, 3, 4},
		"LOW":    []float64{0, 1, 2},
		"CLOSE":  []float64{1, 2, 3},
		"VOLUME": []float64{10, 10, 10},
	}
}

func TestMaiExecutorRun(t *testing.T) {
	ex := NewMaiExecutor()
	source := `//@title 双输出
MA2:MA(CLOSE,2);
FLAG:CROSS(CLOSE,HIGH),COLORRED;
`
	outs, err := ex.Run(source, testBars())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(outs) != 2 {
		t.Fatalf("Run() outputs = %d, want 2", len(outs))
	}
	if outs[0].Name != "MA2" || outs[1].Name != "FLAG" {
		t.Fatalf("output order = %q,%q, want MA2,FLAG", outs[0].Name, outs[1].Name)
	}
	if len(outs[0].Values) != 3 {
		t.Errorf("MA2 values = %v, want length 3", outs[0].Values)
	}
	if outs[1].Kind != KindBool {
		t.Errorf("FLAG kind = %d, want KindBool", outs[1].Kind)
	}
	if len(outs[1].Values) != 3 {
		t.Errorf("FLAG values = %v, want length 3", outs[1].Values)
	}
	if len(outs[1].SuffixParams) != 1 || outs[1].SuffixParams[0] != "COLORRED" {
		t.Errorf("FLAG suffix = %v, want [COLORRED]", outs[1].SuffixParams)
	}
}

func TestMaiExecutorRunCompiled(t *testing.T) {
	ex := NewMaiExecutor()
	program, err := ex.CompileSource("V:MA(CLOSE,2);")
	if err != nil {
		t.Fatalf("CompileSource() error: %v", err)
	}
	first, err := ex.RunCompiled(program, testBars())
	if err != nil {
		t.Fatalf("first RunCompiled() error: %v", err)
	}
	second, err := ex.RunCompiled(program, testBars())
	if err != nil {
		t.Fatalf("second RunCompiled() error: %v", err)
	}
	if len(first) != 1 || len(second) != 1 || first[0].Name != "V" || second[0].Name != "V" {
		t.Fatalf("unexpected outputs: %v / %v", first, second)
	}
}

func TestMaiExecutorRunErrors(t *testing.T) {
	ex := NewMaiExecutor()
	if _, err := ex.Run("X:NOT_EXIST(CLOSE);", testBars()); err == nil {
		t.Error("Run() with undefined function should return error, not panic")
	}
	if _, err := ex.Run("X:MA(CLOSE,2)", testBars()); err == nil {
		t.Error("Run() with syntax error should return error")
	}
}

func TestMaiExecutorResetKeepsFunctions(t *testing.T) {
	ex := NewMaiExecutor()
	ex.Reset()
	if _, err := ex.Run("MA2:MA(CLOSE,2);", testBars()); err != nil {
		t.Fatalf("Reset() lost built-in functions: %v", err)
	}
}

type fakeResolver struct{}

func (f *fakeResolver) ResolveUniverse(sector, indicator string, output int, ctx Ctx) (RankUniverse, string, error) {
	n := len(ctx.Ts)
	mk := func(seq ...float64) []float64 {
		out := make([]float64, n)
		for i := range out {
			out[i] = seq[i%len(seq)]
		}
		return out
	}
	items := []RankItem{
		{Key: "000001", Values: mk(10, 20, 30), Ts: ctx.Ts},
		{Key: "600519", Values: mk(30, 10, 20), Ts: ctx.Ts},
		{Key: "000002", Values: mk(20, 30, 10), Ts: ctx.Ts},
	}
	return RankUniverse{Key: sector + "|" + indicator, Items: items}, ctx.Symbol, nil
}

func assertValues(t *testing.T, name string, got []float64, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s len = %d, want %d (%v)", name, len(got), len(want), got)
	}
	for i := range want {
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got[i]) {
				t.Errorf("%s[%d] = %v, want NaN", name, i, got[i])
			}
			continue
		}
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %v, want %v", name, i, got[i], want[i])
		}
	}
}

func TestINSORTAndINSUM(t *testing.T) {
	ex := NewMaiExecutor()
	ex.SetSectorResolver(&fakeResolver{})
	ctx := Ctx{Symbol: "600519", Ts: []int64{1, 2, 3}, DataVersion: 1, Strict: true}
	outs, err := ex.RunWithContext(
		"R:INSORT('软件服务','KDJ',3,0); S:INSUM('软件服务','KDJ',3,0); M:INSUM('软件服务','KDJ',3,2);",
		ctx, map[string][]float64{"CLOSE": []float64{1, 2, 3}})
	if err != nil {
		t.Fatalf("RunWithContext() error: %v", err)
	}
	if len(outs) != 3 {
		t.Fatalf("outputs = %d, want 3", len(outs))
	}
	assertValues(t, "INSORT", outs[0].Values, []float64{1, 3, 2})
	assertValues(t, "INSUM-sum", outs[1].Values, []float64{60, 60, 60})
	assertValues(t, "INSUM-max", outs[2].Values, []float64{30, 30, 30})
}

type tieResolver struct{}

func (f *tieResolver) ResolveUniverse(sector, indicator string, output int, ctx Ctx) (RankUniverse, string, error) {
	vals := make([]float64, len(ctx.Ts))
	for i := range vals {
		vals[i] = 10
	}
	items := []RankItem{
		{Key: "600519", Values: vals, Ts: ctx.Ts},
		{Key: "000002", Values: vals, Ts: ctx.Ts},
		{Key: "000001", Values: vals, Ts: ctx.Ts},
	}
	return RankUniverse{Key: "tie", Items: items}, ctx.Symbol, nil
}

func TestINSORTTieBreakByCode(t *testing.T) {
	ex := NewMaiExecutor()
	ex.SetSectorResolver(&tieResolver{})
	for symbol, wantRank := range map[string]float64{"000001": 1, "000002": 2, "600519": 3} {
		ctx := Ctx{Symbol: symbol, Ts: []int64{1, 2}, DataVersion: 3, Strict: true}
		outs, err := ex.RunWithContext(
			"R:INSORT('软件服务','KDJ',3,0);",
			ctx, map[string][]float64{"CLOSE": []float64{1, 2}})
		if err != nil {
			t.Fatalf("symbol %s: %v", symbol, err)
		}
		assertValues(t, "tie-"+symbol, outs[0].Values, []float64{wantRank, wantRank})
	}
}

func TestOutputKind(t *testing.T) {
	ex := NewMaiExecutor()
	source := `A:CLOSE;
B:CROSS(CLOSE,OPEN);
买:BUY(CLOSE>OPEN,LOW);
卖:SELL(CLOSE<OPEN,HIGH);`
	outs, err := ex.Run(source, testBars())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	want := []struct {
		name string
		kind OutputKind
	}{
		{"A", KindSeries},
		{"B", KindBool},
		{"买", KindBuy},
		{"卖", KindSell},
	}
	if len(outs) != len(want) {
		t.Fatalf("outputs = %d, want %d", len(outs), len(want))
	}
	for i, w := range want {
		if outs[i].Name != w.name {
			t.Errorf("outs[%d].Name = %q, want %q", i, outs[i].Name, w.name)
		}
		if outs[i].Kind != w.kind {
			t.Errorf("outs[%d].Kind = %d, want %d", i, outs[i].Kind, w.kind)
		}
	}
}
