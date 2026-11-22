package fastcall

import (
	"testing"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// BenchmarkInvoker 去反射（类型化调用器）。
func BenchmarkInvoker(t *testing.B) {
	inv, ok := Invoker("EMA")
	if !ok {
		t.Fatal("EMA 未覆盖")
	}
	args := []any{seq(500), 12}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := inv(args); err != nil {
			t.Fatal(err)
		}
	}
}

// BenchmarkReflect 反射调用（原路径）。
func BenchmarkReflect(t *testing.B) {
	args := []any{seq(500), 12}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if _, err := indicators.CallIndicatorByReflection("EMA", args); err != nil {
			t.Fatal(err)
		}
	}
}
