package vm

import "github.com/lyr-2000/mylang/pkg/mylang"

// Compile 把 AST 编译成字节码（后序遍历，生成线性指令序列）。
// mylang 是声明式语言（无循环/分支），因此字节码是纯线性的，无跳转指令。
func Compile(prog *mylang.Program) *Chunk {
	c := &compiler{
		chunk:    &Chunk{},
		nameMap:  map[string]int{},
		constMap: map[interface{}]int{},
	}
	for _, stmt := range prog.Statements {
		c.compileStmt(stmt)
	}
	c.emit(OpReturn, 0, 0)
	return c.chunk
}

type compiler struct {
	chunk    *Chunk
	nameMap  map[string]int
	constMap map[interface{}]int
}

func (c *compiler) compileStmt(stmt mylang.Statement) {
	switch s := stmt.(type) {
	case *mylang.AssignmentStatement:
		c.compileExpr(s.Value)
		if s.IsOutputVar {
			c.emit(OpStoreOutput, c.nameIndex(s.Name.Value), 0)
		} else {
			c.emit(OpStoreVar, c.nameIndex(s.Name.Value), 0)
		}
	case *mylang.ExpressionStatement:
		c.compileExpr(s.Expression)
		c.emit(OpPop, 0, 0)
	}
}

func (c *compiler) compileExpr(e mylang.Expression) {
	switch n := e.(type) {
	case *mylang.NumberLiteral:
		c.emit(OpLoadConst, c.constIndex(n.Value), 0)
	case *mylang.StringLiteral:
		c.emit(OpLoadConst, c.constIndex(n.Value), 0)
	case *mylang.Identifier:
		c.emit(OpLoadVar, c.nameIndex(n.Value), 0)
	case *mylang.BinaryExpression:
		c.compileExpr(n.Left)
		c.compileExpr(n.Right)
		c.emit(binOpcode(n.Operator), 0, 0)
	case *mylang.UnaryExpression:
		c.compileExpr(n.Right)
		switch n.Operator {
		case "-":
			c.emit(OpNeg, 0, 0)
		case "!", "NOT":
			c.emit(OpNot, 0, 0)
		}
	case *mylang.FunctionCall:
		for _, a := range n.Arguments {
			c.compileExpr(a)
		}
		name := ""
		if id, ok := n.Function.(*mylang.Identifier); ok {
			name = id.Value
		} else {
			name = n.Function.String()
		}
		c.emit(OpCall, c.nameIndex(name), len(n.Arguments))
	}
}

// binOpcode 把 AST 的 operator 字符串映射成字节码操作码。
func binOpcode(op string) Opcode {
	switch op {
	case "+":
		return OpAdd
	case "-":
		return OpSub
	case "*":
		return OpMul
	case "/":
		return OpDiv
	case ">":
		return OpGT
	case "<":
		return OpLT
	case ">=":
		return OpGE
	case "<=":
		return OpLE
	case "==", "=":
		return OpEQ
	case "!=":
		return OpNE
	case "AND":
		return OpAnd
	case "OR", "or":
		return OpOr
	}
	return OpAdd
}

func (c *compiler) emit(op Opcode, idx, n int) {
	c.chunk.Code = append(c.chunk.Code, Instr{Op: op, Idx: idx, N: n})
}

func (c *compiler) nameIndex(name string) int {
	if i, ok := c.nameMap[name]; ok {
		return i
	}
	i := len(c.chunk.Names)
	c.chunk.Names = append(c.chunk.Names, name)
	c.nameMap[name] = i
	return i
}

func (c *compiler) constIndex(v interface{}) int {
	if i, ok := c.constMap[v]; ok {
		return i
	}
	i := len(c.chunk.Consts)
	c.chunk.Consts = append(c.chunk.Consts, v)
	c.constMap[v] = i
	return i
}
