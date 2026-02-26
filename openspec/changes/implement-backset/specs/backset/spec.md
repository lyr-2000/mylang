## ADDED Requirements

### Requirement: BACKSET 回填语义

BACKSET(S, N) SHALL 接受条件序列 S（[]bool）与周期 N（int 或 []int），返回与 S 等长的 []bool。当 S[i] 为 true 时，SHALL 将结果中从 max(0, i-N+1) 到 i（含）的下标置为 true；N 表示“当前及前 N-1 根”共 N 根 K 线。

#### Scenario: 标量 N=2 收阳回填

- **WHEN** S = [false, true, false]，N = 2
- **THEN** 结果 [1]=true 导致 [0,1] 置为 true，故结果为 [true, true, false]

#### Scenario: 仅当前 bar 置位

- **WHEN** S = [false, true, false]，N = 1
- **THEN** 仅 [1] 置为 true，结果为 [false, true, false]

#### Scenario: 结果长度与 S 一致

- **WHEN** S 长度为 L（任意非负）
- **THEN** 返回切片长度 SHALL 等于 L

#### Scenario: N 为序列按索引取值

- **WHEN** S = [false, true, true]，N = [1, 2, 1]（与 S 同长）
- **THEN** i=1 时 N=2 回填 [0,1]；i=2 时 N=1 仅 [2]；结果为 [true, true, true]

#### Scenario: N 序列较短时默认 1

- **WHEN** S = [false, true]，N = []int{5}（仅一个元素）
- **THEN** 在 i=1 时 N 取 N[1] 不存在，SHALL 使用默认 1，仅 [1] 为 true，结果为 [false, true]

#### Scenario: N≤0 视为 1

- **WHEN** S = [true]，N = 0 或负数
- **THEN** SHALL 将 N 视为 1，仅当前 bar 置为 true，结果为 [true]
