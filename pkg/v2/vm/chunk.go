// Package vm 实现麦语言的字节码虚拟机（零侵入原包）。
//
// 原 mylang 是「树遍历解释器」：编译止步于 AST，执行时用 Eval 递归 switch 分发。
// 本包把 AST 编译成线性字节码，执行时用「取指-解码-执行」循环替代递归 + switch，
// 消除每条语句的树遍历开销；列式运算（数组逐元素）保留，并顺手把 operator 的
// switch 提到循环体外（消除每次迭代的字符串比较）。
//
// 不 import api 包（避免循环依赖）：函数表由调用方通过 SetFunc 注入，
// 变量表通过 Run 的 vars 参数注入。
package vm

// Opcode 字节码操作码。
type Opcode uint8

const (
	OpLoadConst   Opcode = iota // 压常量（Idx=常量池索引）
	OpLoadVar                   // 压变量（Idx=名称池索引）
	OpStoreVar                  // 存中间变量（Idx=名称池索引）
	OpStoreOutput               // 存输出变量（Idx=名称池索引）
	OpPop                       // 丢弃栈顶（表达式语句）
	OpAdd                       // 二元：+
	OpSub                       // 二元：-
	OpMul                       // 二元：*
	OpDiv                       // 二元：/
	OpGT                        // 二元：>
	OpLT                        // 二元：<
	OpGE                        // 二元：>=
	OpLE                        // 二元：<=
	OpEQ                        // 二元：==
	OpNE                        // 二元：!=
	OpAnd                       // 二元：AND
	OpOr                        // 二元：OR
	OpNot                       // 一元：!
	OpNeg                       // 一元：-
	OpCall                      // 函数调用（Idx=名称池索引，N=参数个数）
	OpReturn                    // 结束
)

// Instr 一条字节码指令。
type Instr struct {
	Op  Opcode
	Idx int // 常量池 / 名称池索引
	N   int // OpCall 的参数个数
}

// Chunk 一段编译产物：线性指令 + 常量池 + 名称池（变量名/函数名共享）。
type Chunk struct {
	Code   []Instr
	Consts []interface{}
	Names  []string
}
