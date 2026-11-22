package api

import (
	"sync"

	"github.com/lyr-2000/mylang/pkg/mylang"
)

// Pool 并发安全的 MaiExecutor 池。
//
// 预编译一次 Program（只读、跨 goroutine 共享），通过 sync.Pool 池化 MaiExecutor 实例。
// 每个 goroutine 通过 Run / RunWithContext 从池里取一个独立实例执行，用完放回，
// 既避免并发共用同一个有状态的解释器，又避免每次 NewMaiExecutor 的注册开销。
//
// 适用场景：脚本固定、按不同数据批量执行（全市场回测、逐股票算指标）。
// 注意：脚本不同请各建一个 Pool；一个 Pool 只对应一份预编译 Program。
type Pool struct {
	program *mylang.Program
	newFn   func() *MaiExecutor
	pool    sync.Pool
}

// NewPool 构造池并预编译 source。opts 可选，在创建每个 MaiExecutor 时应用
// （如 SetSectorResolver / SetCustomVariableGetter / SetDateTimeKey）。
func NewPool(source string, opts ...func(*MaiExecutor)) (*Pool, error) {
	e := NewMaiExecutor()
	for _, o := range opts {
		o(e)
	}
	program, err := e.CompileSource(source)
	if err != nil {
		return nil, err
	}

	p := &Pool{
		program: program,
		newFn: func() *MaiExecutor {
			ne := NewMaiExecutor()
			for _, o := range opts {
				o(ne)
			}
			return ne
		},
	}
	p.pool.New = func() any { return p.newFn() }
	return p, nil
}

// Program 返回池内共享的预编译 Program（只读）。
func (p *Pool) Program() *mylang.Program {
	return p.program
}

// Run 从池取实例执行预编译脚本，注入 vars，返回有序输出。
func (p *Pool) Run(vars map[string][]float64) ([]Output, error) {
	e := p.pool.Get().(*MaiExecutor)
	defer p.pool.Put(e)
	return e.RunCompiled(p.program, vars)
}

// RunWithContext 带上下文执行（INSORT/INSUM 等横截面函数依赖 ctx.Symbol/Ts/DataVersion）。
func (p *Pool) RunWithContext(ctx Ctx, vars map[string][]float64) ([]Output, error) {
	e := p.pool.Get().(*MaiExecutor)
	defer p.pool.Put(e)
	return e.RunCompiledWithContext(p.program, ctx, vars)
}
