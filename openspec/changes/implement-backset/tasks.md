## 1. 实现 BACKSET 函数体

- [x] 1.1 将 BACKSET 签名定为 `func BACKSET(S []bool, N any) []bool`，在 MyTT.go 中补全
- [x] 1.2 实现 N 为 int/float64 的标量分支：对每个 i 若 S[i] 为 true，将 [max(0,i-n+1), i] 置为 true
- [x] 1.3 实现 N 为 []int 的序列分支：按索引取 n=N[i]，若 i>=len(N) 或 n<=0 则 n=1，再按相同区间规则回填
- [x] 1.4 使用 cast.ToInt 支持 N 为 float64，与 REF 行为一致；空 S 时返回 nil 或零长度切片

## 2. 边界与一致性

- [x] 2.1 保证返回长度始终为 len(S)；N≤0 时按 1 处理
- [x] 2.2 确认与 EVERY/EXIST/FILTER 等同文件函数的风格一致（注释、命名、无额外依赖）
