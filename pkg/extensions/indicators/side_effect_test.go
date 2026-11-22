package indicators

import "testing"

func TestIsPureBuiltin(t *testing.T) {
	// 内置指标函数在 init 自动登记为纯函数
	for _, name := range []string{"EMA", "MA", "REF", "CROSS", "BUY", "CONST"} {
		if !IsPure(name) {
			t.Errorf("内置函数 IsPure(%q) = false, want true", name)
		}
	}
}

func TestIsPureDefaultFalse(t *testing.T) {
	// 未登记的（context 函数 / 用户扩展函数）默认不纯
	for _, name := range []string{"INSORT", "INSUM", "MyCustomFunc"} {
		if IsPure(name) {
			t.Errorf("未登记函数 IsPure(%q) = true, want false（应默认有副作用）", name)
		}
	}
}

func TestMarkPure(t *testing.T) {
	// 显式声明后为纯
	const name = "MyOptInFunc"
	if IsPure(name) {
		t.Fatalf("前置条件失败：%q 不应已登记", name)
	}
	MarkPure(name)
	if !IsPure(name) {
		t.Errorf("MarkPure 后 IsPure(%q) = false, want true", name)
	}
}
