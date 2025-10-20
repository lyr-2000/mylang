package api

import (
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