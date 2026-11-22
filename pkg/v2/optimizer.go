// Package v2 是 mylang 的「优化版」孵化包。
//
// 设计目标：不动 pkg/mylang、pkg/api、pkg/extensions 的现有实现，把编译期优化
// （常量折叠、公共子表达式消除）放在本包，作为独立 AST 优化 pass。验证收益后，
// 再由宿主在 api.New 入口处切换到 v2 的执行器。
//
// 优化只做「零风险」的 AST 变换：
//   - 常量折叠：纯常量子树（数字字面量的算术/比较）在编译期预计算。
//   - CSE：结构相同的「叶子参数函数调用」消除重复计算，替换成中间变量引用。
//
// 刻意排除有状态/横截面函数（INSORT/INSUM）与非叶子调用，避免改变求值时机引入
// 语义差异。Program 由 CompileSource 刚产出、尚未被其他执行器共享，故原地优化安全。
package v2

import (
	"fmt"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/mylang"
)

// Optimize 对 Program 做 AST 优化（常量折叠 + CSE），返回优化后的 Program。
// 有语法错误时直接原样返回，不做变换。
func Optimize(p *mylang.Program) *mylang.Program {
	if p == nil || len(p.Errors) > 0 {
		return p
	}
	for _, stmt := range p.Statements {
		if as, ok := stmt.(*mylang.AssignmentStatement); ok && as.Value != nil {
			as.Value = constFold(as.Value)
		}
	}
	p.Statements = cse(p.Statements)
	p.Statements = dce(p.Statements)
	return p
}

// ---- 常量折叠 ----

// constFold 递归折叠纯常量子树，返回（可能的）新表达式。
func constFold(e mylang.Expression) mylang.Expression {
	switch n := e.(type) {
	case *mylang.BinaryExpression:
		n.Left = constFold(n.Left)
		n.Right = constFold(n.Right)
		l, lok := n.Left.(*mylang.NumberLiteral)
		r, rok := n.Right.(*mylang.NumberLiteral)
		if lok && rok {
			if v, ok := foldBinary(l.Value, n.Operator, r.Value); ok {
				return &mylang.NumberLiteral{Token: n.Token, Value: v}
			}
		}
		return n
	case *mylang.UnaryExpression:
		n.Right = constFold(n.Right)
		if r, ok := n.Right.(*mylang.NumberLiteral); ok {
			switch n.Operator {
			case "-":
				return &mylang.NumberLiteral{Token: n.Token, Value: -r.Value}
			case "+":
				return &mylang.NumberLiteral{Token: n.Token, Value: r.Value}
			}
		}
		return n
	case *mylang.FunctionCall:
		for i, a := range n.Arguments {
			n.Arguments[i] = constFold(a)
		}
		return n
	default:
		return e
	}
}

// foldBinary 对两个数字字面量做算术/比较折叠；ok=false 表示不可折叠（如除零、未知运算符）。
func foldBinary(l float64, op string, r float64) (float64, bool) {
	switch op {
	case "+":
		return l + r, true
	case "-":
		return l - r, true
	case "*":
		return l * r, true
	case "/":
		if r == 0 {
			return 0, false
		}
		return l / r, true
	case ">":
		return boolNum(l > r), true
	case "<":
		return boolNum(l < r), true
	case ">=":
		return boolNum(l >= r), true
	case "<=":
		return boolNum(l <= r), true
	case "==", "=":
		return boolNum(l == r), true
	case "!=":
		return boolNum(l != r), true
	}
	return 0, false
}

func boolNum(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// ---- 公共子表达式消除（CSE）----

// cse 对「结构相同的叶子参数函数调用」消除重复计算。
// 第一次出现时插入中间变量赋值（__cseN := call），后续出现替换成对中间变量的引用。
func cse(stmts []mylang.Statement) []mylang.Statement {
	// 第一遍：统计每个可 CSE 调用的出现次数。
	counts := map[string]int{}
	for _, stmt := range stmts {
		as, ok := stmt.(*mylang.AssignmentStatement)
		if !ok {
			continue
		}
		walkCalls(as.Value, func(fc *mylang.FunctionCall) {
			if isCSEable(fc) {
				counts[fc.String()]++
			}
		})
	}

	// 第二遍：替换重复调用，插入中间变量赋值。
	seen := map[string]string{} // hash -> 中间变量名
	tmpIdx := 0
	out := make([]mylang.Statement, 0, len(stmts)+len(counts))
	for _, stmt := range stmts {
		as, ok := stmt.(*mylang.AssignmentStatement)
		if !ok {
			out = append(out, stmt)
			continue
		}
		var pending []mylang.Statement
		as.Value = replaceDup(as.Value, counts, seen, &tmpIdx, &pending)
		out = append(out, pending...)
		out = append(out, as)
	}
	return out
}

// isCSEable 只对「纯函数白名单内且参数全为叶子」的函数调用做 CSE。
// 纯函数白名单由 indicators 包声明（IsPure）；不在白名单（含扩展函数）默认视为
// 有副作用不做优化，非叶子参数也排除（避免替换导致 hash 漂移）。
func isCSEable(fc *mylang.FunctionCall) bool {
	if !indicators.IsPure(fc.Function.String()) {
		return false
	}
	for _, a := range fc.Arguments {
		switch a.(type) {
		case *mylang.Identifier, *mylang.NumberLiteral:
			// 叶子参数，允许
		default:
			return false
		}
	}
	return true
}

// walkCalls 遍历表达式的所有函数调用（含嵌套）。
func walkCalls(e mylang.Expression, fn func(*mylang.FunctionCall)) {
	switch n := e.(type) {
	case *mylang.FunctionCall:
		fn(n)
		for _, a := range n.Arguments {
			walkCalls(a, fn)
		}
	case *mylang.BinaryExpression:
		walkCalls(n.Left, fn)
		walkCalls(n.Right, fn)
	case *mylang.UnaryExpression:
		walkCalls(n.Right, fn)
	}
}

// replaceDup 替换重复调用：首次出现插入 __cseN := call，后续出现替换成 __cseN。
func replaceDup(e mylang.Expression, counts map[string]int, seen map[string]string,
	tmpIdx *int, pending *[]mylang.Statement) mylang.Expression {

	switch n := e.(type) {
	case *mylang.FunctionCall:
		for i, a := range n.Arguments {
			n.Arguments[i] = replaceDup(a, counts, seen, tmpIdx, pending)
		}
		if !isCSEable(n) {
			return n
		}
		h := n.String()
		if counts[h] <= 1 {
			return n
		}
		if name, ok := seen[h]; ok {
			return &mylang.Identifier{Token: n.Token, Value: name}
		}
		*tmpIdx++
		name := fmt.Sprintf("__cse%d", *tmpIdx)
		seen[h] = name
		*pending = append(*pending, &mylang.AssignmentStatement{
			Token:       n.Token,
			Name:        &mylang.Identifier{Token: n.Token, Value: name},
			Value:       n,
			IsOutputVar: false,
		})
		return &mylang.Identifier{Token: n.Token, Value: name}
	case *mylang.BinaryExpression:
		n.Left = replaceDup(n.Left, counts, seen, tmpIdx, pending)
		n.Right = replaceDup(n.Right, counts, seen, tmpIdx, pending)
		return n
	case *mylang.UnaryExpression:
		n.Right = replaceDup(n.Right, counts, seen, tmpIdx, pending)
		return n
	default:
		return e
	}
}

// ---- 死代码消除（DCE）----

// dce 删除「未被输出变量（间接）引用」的中间变量赋值。
// 从输出变量（":" 赋值）出发反向追踪依赖，删除不在存活集里的中间变量（":=" 赋值）。
// 有重复赋值（同一变量名多次 ":="）时保守跳过，避免误删。
func dce(stmts []mylang.Statement) []mylang.Statement {
	type info struct {
		stmt *mylang.AssignmentStatement
		deps map[string]struct{}
	}
	var assigns []info
	outputs := map[string]struct{}{}
	seenAssign := map[string]bool{}

	for _, stmt := range stmts {
		as, ok := stmt.(*mylang.AssignmentStatement)
		if !ok {
			continue
		}
		if seenAssign[as.Name.Value] {
			return stmts // 重复赋值，保守跳过
		}
		seenAssign[as.Name.Value] = true
		deps := map[string]struct{}{}
		collectRefs(as.Value, deps)
		assigns = append(assigns, info{stmt: as, deps: deps})
		if as.IsOutputVar {
			outputs[as.Name.Value] = struct{}{}
		}
	}

	// 从输出变量反向迭代到不动点，求存活变量。
	live := map[string]struct{}{}
	for name := range outputs {
		live[name] = struct{}{}
	}
	for changed := true; changed; {
		changed = false
		for _, a := range assigns {
			if _, ok := live[a.stmt.Name.Value]; !ok {
				continue
			}
			for dep := range a.deps {
				if _, ok := live[dep]; !ok {
					live[dep] = struct{}{}
					changed = true
				}
			}
		}
	}

	out := make([]mylang.Statement, 0, len(assigns))
	for _, a := range assigns {
		if a.stmt.IsOutputVar {
			out = append(out, a.stmt)
			continue
		}
		if _, ok := live[a.stmt.Name.Value]; ok {
			out = append(out, a.stmt)
		}
	}
	return out
}

// collectRefs 收集表达式里的「变量引用」（排除函数名）。
func collectRefs(e mylang.Expression, refs map[string]struct{}) {
	switch n := e.(type) {
	case *mylang.Identifier:
		refs[n.Value] = struct{}{}
	case *mylang.FunctionCall:
		for _, a := range n.Arguments {
			collectRefs(a, refs)
		}
	case *mylang.BinaryExpression:
		collectRefs(n.Left, refs)
		collectRefs(n.Right, refs)
	case *mylang.UnaryExpression:
		collectRefs(n.Right, refs)
	}
}
