## Why

MyTT 指标库需要补齐通达信公式中的 BACKSET 未来函数，以便在回测或指标表达中实现“满足条件时向前回填若干周期为 1”的语义；当前 `pkg/extensions/indicators/MyTT.go` 中仅有空壳 `func BACKSET() []bool`，需要按注释规范完成实现。

## What Changes

- 在 `pkg/extensions/indicators/MyTT.go` 中实现 `BACKSET(S []bool, N []int) []bool`（或等价签名），行为与通达信 BACKSET 一致。
- 语义：当 `S[i]` 为 true 时，将结果序列中从 `max(0, i-N+1)` 到 `i`（含）的下标置为 true；N 可为标量或按索引的序列，表示“当前及前 N-1 根”共 N 根 K 线。
- 边界：结果长度以 `len(S)` 为准；N≤0 或缺失时按 1 处理；支持 N 为 `[]int` 时按索引取 `N[i]`，超出部分用默认 1。

## Capabilities

### New Capabilities

- `backset`: 通达信 BACKSET 函数在 MyTT 指标库中的实现规范与行为约定（入参、返回值、边界与示例）。

### Modified Capabilities

- （无）

## Impact

- **代码**：仅影响 `pkg/extensions/indicators/MyTT.go`，新增/补全 BACKSET 函数体，不改变其他导出 API。
- **依赖**：无新增依赖。
- **调用方**：任何使用 MyTT 且需要 BACKSET 语义的公式或回测逻辑将可调用该函数。
