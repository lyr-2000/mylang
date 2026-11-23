package membuf

import (
	"testing"
	"unsafe"
)

func ptr(s []float64) uintptr {
	return uintptr(unsafe.Pointer(&s[0]))
}

// TestPoolReuse 验证 Put 后 Get 能复用同一底层数组。
func TestPoolReuse(t *testing.T) {
	p := New()
	a := p.Get(500)
	a[0] = 123
	a[499] = 456
	pa := ptr(a)
	p.Put(a)

	b := p.Get(500)
	if len(b) != 500 {
		t.Fatalf("Get(500) 长度 %d, want 500", len(b))
	}
	if ptr(b) != pa {
		t.Fatal("Put 后 Get 未复用同一底层数组")
	}
	// 复用后数据应保留（buffer 未清零）
	if b[0] != 123 || b[499] != 456 {
		t.Fatalf("复用后数据异常: b[0]=%v b[499]=%v", b[0], b[499])
	}
}

// TestPoolBuckets 验证不同长度落到不同档位（容量被档位容量对齐）。
func TestPoolBuckets(t *testing.T) {
	p := New()
	a := p.Get(100) // 128 档
	if cap(a) != 128 {
		t.Fatalf("Get(100) cap=%d, want 128", cap(a))
	}
	p.Put(a)

	b := p.Get(300) // 512 档
	if cap(b) != 512 {
		t.Fatalf("Get(300) cap=%d, want 512", cap(b))
	}
	p.Put(b)

	// 再次 Get(100) 应复用到 128 档的 a
	c := p.Get(100)
	if ptr(c) != ptr(a) {
		t.Fatal("Get(100) 未复用 128 档 buffer")
	}
}

// TestPoolOversize 验证超过最大档位时不池化（直接新分配，Put 丢弃）。
func TestPoolOversize(t *testing.T) {
	p := New()
	s := p.Get(100000)
	if cap(s) != 100000 {
		t.Fatalf("超大 Get cap=%d, want 100000", cap(s))
	}
	p.Put(s) // 超档，应被丢弃（不 panic 即可）
}

// TestPoolGetZero 验证零长度不 panic。
func TestPoolGetZero(t *testing.T) {
	p := New()
	if s := p.Get(0); len(s) != 0 {
		t.Fatalf("Get(0) 长度 %d, want 0", len(s))
	}
}
