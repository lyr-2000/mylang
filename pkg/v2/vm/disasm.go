package vm

import (
	"fmt"
	"strings"
)

// Disassemble 把字节码反汇编成可读的多行文本（用于调试/查看编译产物）。
func (c *Chunk) Disassemble() string {
	var sb strings.Builder
	for pc, instr := range c.Code {
		sb.WriteString(fmt.Sprintf("%04d  %-14s", pc, opcodeName(instr.Op)))
		switch instr.Op {
		case OpLoadConst:
			sb.WriteString(fmt.Sprintf("%v", c.Consts[instr.Idx]))
		case OpLoadVar, OpStoreVar, OpStoreOutput:
			sb.WriteString(c.Names[instr.Idx])
		case OpCall:
			sb.WriteString(fmt.Sprintf("%s  (args=%d)", c.Names[instr.Idx], instr.N))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func opcodeName(op Opcode) string {
	switch op {
	case OpLoadConst:
		return "LOAD_CONST"
	case OpLoadVar:
		return "LOAD_VAR"
	case OpStoreVar:
		return "STORE_VAR"
	case OpStoreOutput:
		return "STORE_OUTPUT"
	case OpPop:
		return "POP"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpGT:
		return "GT"
	case OpLT:
		return "LT"
	case OpGE:
		return "GE"
	case OpLE:
		return "LE"
	case OpEQ:
		return "EQ"
	case OpNE:
		return "NE"
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpNot:
		return "NOT"
	case OpNeg:
		return "NEG"
	case OpCall:
		return "CALL"
	case OpReturn:
		return "RETURN"
	}
	return "UNKNOWN"
}
