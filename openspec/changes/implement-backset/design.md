## Context

- MyTT 位于 `pkg/extensions/indicators/MyTT.go`，与 EVERY、EXIST、FILTER 等一致：入参为 `S []bool` 时，结果长度以 `len(S)` 为准；标量 N 通常为 `int`，序列 N 为 `[]int` 或通过 `any` 支持多类型（参见 REF）。
- 当前 BACKSET 仅有空壳 `func BACKSET() []bool`，需补全签名与实现，且与文件内现有风格（单文件、无额外依赖、纯逻辑）保持一致。

## Goals / Non-Goals

**Goals:**

- 实现 BACKSET(S, N)，语义与通达信一致：S[i]==true 时，将 [i-N+1, i] 区间置为 true。
- 支持 N 为标量（int）或序列（[]int）；序列时按索引取 N[i]，缺省为 1。
- 与同文件内 EVERY/EXIST/FILTER 等函数的长度与边界约定一致。

**Non-Goals:**

- 不改变其他 MyTT 函数；不引入新依赖；不提供 Python/其他语言绑定。

## Decisions

1. **签名形式**  
   - 采用 `func BACKSET(S []bool, N any) []bool`，与 REF(S, N) 类似，内部用类型断言区分 `int` 与 `[]int`，便于调用方既写 `BACKSET(S, 2)` 又写 `BACKSET(S, someSlice)`。  
   - 备选：仅 `BACKSET(S []bool, N []int)`，标量由调用方构造 `[]int{N,...,N}`；未采用因调用体验不如 any 灵活。

2. **N 的默认值**  
   - 当 N 为序列且 `i >= len(N)` 时使用 1；当 N≤0 或转换失败时也视为 1（只影响当前 bar）。与“若干周期”的直觉一致，且避免 panic。

3. **区间与边界**  
   - 对每个 i 满足 S[i]==true：`start = max(0, i - n + 1)`，对 j ∈ [start, i] 设置 result[j]=true。不分配临时切片，单层循环按 i 递增即可覆盖所有需回填的区间。

## Risks / Trade-offs

- **未来函数语义**：BACKSET 会“用当前信息改过去”，回测时若误用会导致前视偏差；文档注释已说明为未来函数，不在本实现中额外限制。
- **N 为 float64**：若调用方传入 float64，可沿用 REF 的做法用 `cast.ToInt(N)` 统一成 int，避免 panic。

## Migration Plan

- 仅新增/补全函数体，无迁移步骤；若曾存在错误签名，调用方需改为传入 (S, N) 两参。

## Open Questions

- 无。
