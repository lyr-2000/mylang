package api

import "testing"

// TestProtectSystemFunction 验证：注册外部函数时禁止覆盖系统函数（内置指标函数 + 上下文函数）。
func TestProtectSystemFunction(t *testing.T) {
	ex := NewMaiExecutor()

	cases := []struct {
		name string
		fn   func()
	}{
		{"内置指标函数 EMA", func() { ex.RegisterFunction("EMA", func([]interface{}) interface{} { return nil }) }},
		{"上下文函数 INSORT", func() { ex.RegisterContextFunction("INSORT", nil) }},
		{"上下文函数 INSUM", func() { ex.RegisterContextFunction("INSUM", nil) }},
	}

	for _, c := range cases {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s：应 panic（禁止覆盖系统函数），但未触发", c.name)
				}
			}()
			c.fn()
		}()
	}
}

// TestRegisterUserFunc 验证：注册非系统函数名正常，不影响覆盖保护。
func TestRegisterUserFunc(t *testing.T) {
	ex := NewMaiExecutor()
	ex.RegisterFunction("MyCustomFunc", func([]interface{}) interface{} { return nil })
}
