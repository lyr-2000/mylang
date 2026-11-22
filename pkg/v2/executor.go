package v2

import (
	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/mylang"
)

// Executor 是带 AST 优化的执行器：复用原包 api.MaiExecutor 的解析与执行，
// 在编译产物上叠加 v2 的优化 pass。核心方法对齐 api.MaiExecutor，验证收益后
// 可直接在 api.New 入口处切换到本实现。
type Executor struct {
	inner *api.MaiExecutor
}

// New 构造优化执行器。
func New() *Executor {
	return &Executor{inner: api.NewMaiExecutor()}
}

// Wrap 用已有执行器构造（复用其注册的自定义函数 / SectorResolver 等配置）。
func Wrap(inner *api.MaiExecutor) *Executor {
	return &Executor{inner: inner}
}

// Inner 返回底层执行器，供需要访问解释器高级能力的场景。
func (e *Executor) Inner() *api.MaiExecutor { return e.inner }

// Compile 编译并优化脚本，返回优化后的 AST（可缓存，配合 RunCompiled 复用）。
func (e *Executor) Compile(source string) (*mylang.Program, error) {
	prog, err := e.inner.CompileSource(source)
	if err != nil {
		return nil, err
	}
	return Optimize(prog), nil
}

// Run 编译 + 优化 + 执行，返回有序输出。
func (e *Executor) Run(source string, vars map[string][]float64) ([]api.Output, error) {
	prog, err := e.Compile(source)
	if err != nil {
		return nil, err
	}
	return e.inner.RunCompiled(prog, vars)
}

// RunWithContext 带上下文运行（INSORT/INSUM 需要 ctx）。
func (e *Executor) RunWithContext(source string, ctx api.Ctx, vars map[string][]float64) ([]api.Output, error) {
	prog, err := e.Compile(source)
	if err != nil {
		return nil, err
	}
	return e.inner.RunCompiledWithContext(prog, ctx, vars)
}

// RunCompiled 执行外部缓存（需已优化）的 Program。
func (e *Executor) RunCompiled(prog *mylang.Program, vars map[string][]float64) ([]api.Output, error) {
	return e.inner.RunCompiled(prog, vars)
}

// RunCompiledWithContext 带上下文执行外部缓存 Program。
func (e *Executor) RunCompiledWithContext(prog *mylang.Program, ctx api.Ctx, vars map[string][]float64) ([]api.Output, error) {
	return e.inner.RunCompiledWithContext(prog, ctx, vars)
}
