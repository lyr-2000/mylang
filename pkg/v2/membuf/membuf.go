package membuf

import "sync"

// Pool 分档复用 []float64 底层数组，降低列式运算中间结果的 GC 压力。
//
// 通过「长度分档 + sync.Pool」实现：Get(n) 从容量 >= n 的档位复用空闲数组，
// Put(s) 按 cap(s) 放回对应档位。只复用「容量正好等于档位」的 buffer，避免长度错配。
//
// 非严格并发安全（每个 VM 实例独占一个 Pool），但内部用 sync.Pool 保证安全。
type Pool struct {
	sizes   []int
	buckets []sync.Pool
}

// New 创建默认分档（128/256/512/1024/2048/4096/8192/16384）的池。
func New() *Pool {
	sizes := []int{128, 256, 512, 1024, 2048, 4096, 8192, 16384}
	p := &Pool{sizes: sizes, buckets: make([]sync.Pool, len(sizes))}
	return p
}

// Get 返回长度 n、容量 >= n 的 []float64；命中空闲 buffer 时复用底层数组，否则按档位容量新分配。
func (p *Pool) Get(n int) []float64 {
	if n <= 0 {
		return make([]float64, n)
	}
	idx, ok := p.bucketIndex(n)
	if !ok {
		return make([]float64, n) // 超过最大档位，直接按 n 分配（不池化）
	}
	if v, ok := p.buckets[idx].Get().([]float64); ok {
		return v[:n]
	}
	return make([]float64, n, p.sizes[idx]) // 新分配，cap 对齐档位容量
}

// Put 归还 buffer。只接受「容量正好等于档位」的 buffer（即本池 Get 产出的），其余丢弃交给 GC。
func (p *Pool) Put(s []float64) {
	if len(s) == 0 {
		return
	}
	idx, ok := p.bucketIndex(cap(s))
	if !ok {
		return
	}
	if cap(s) == p.sizes[idx] {
		p.buckets[idx].Put(s)
	}
}

// bucketIndex 返回「容量 >= n 的最小档位」下标。
func (p *Pool) bucketIndex(n int) (int, bool) {
	for i, sz := range p.sizes {
		if n <= sz {
			return i, true
		}
	}
	return 0, false
}
