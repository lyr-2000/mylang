package mylang

import (
	"fmt"
	"strconv"
	"strings"
)

// Node 代表语法树中的一个节点
type Node interface {
	String() string
}

// Statement 代表一个语句
type Statement interface {
	Node
	statementNode()
}

// Expression 代表一个表达式
type Expression interface {
	Node
	expressionNode()
}

// Program 代表整个程序
type Program struct {
	Statements []Statement
	Errors     []string // 存储语法错误
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// Identifier 代表一个标识符
type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }

// NumberLiteral 代表一个数字字面量
type NumberLiteral struct {
	Token Token
	Value float64
}

func (nl *NumberLiteral) expressionNode() {}
func (nl *NumberLiteral) String() string  { return nl.Token.Literal }

// StringLiteral 代表一个字符串字面量
type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string  { return sl.Token.Literal }

// BinaryExpression 代表一个二元表达式
type BinaryExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) String() string {
	leftStr := "<nil>"
	rightStr := "<nil>"
	if be.Left != nil {
		leftStr = be.Left.String()
	}
	if be.Right != nil {
		rightStr = be.Right.String()
	}
	return "(" + leftStr + " " + be.Operator + " " + rightStr + ")"
}

// UnaryExpression 代表一个一元表达式
type UnaryExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) expressionNode() {}
func (ue *UnaryExpression) String() string {
	rightStr := "<nil>"
	if ue.Right != nil {
		rightStr = ue.Right.String()
	}
	return "(" + ue.Operator + " " + rightStr + ")"
}

// FunctionCall 代表一个函数调用
type FunctionCall struct {
	Token     Token
	Function  Expression
	Arguments []Expression
}

func (fc *FunctionCall) expressionNode() {}
func (fc *FunctionCall) String() string {
	args := []string{}
	for _, a := range fc.Arguments {
		if a != nil {
			args = append(args, a.String())
		} else {
			args = append(args, "<nil>")
		}
	}
	functionStr := "<nil>"
	if fc.Function != nil {
		functionStr = fc.Function.String()
	}
	return functionStr + "(" + strings.Join(args, ", ") + ")"
}

// AssignmentStatement 代表一个赋值语句
type AssignmentStatement struct {
	Token        Token
	Name         *Identifier
	Value        Expression
	IsOutputVar  bool     // true表示是画图变量赋值(:), false表示普通赋值(:=)
	SuffixParams []string // 存储修饰符，如 COLORRED, NODRAW
	// IsDrawingStatement 标记「无变量名画图语句」（如 DRAWTEXT_FIX(...),COLORRED;）。
	// 其输出变量名取函数名，执行时遇到重名会追加序号，避免多条同名画图语句互相覆盖。
	IsDrawingStatement bool
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) String() string {
	result := ""
	if as.IsOutputVar {
		result = as.Name.String() + " : " + as.Value.String()
	} else {
		result = as.Name.String() + " := " + as.Value.String()
	}

	// 添加修饰符
	if len(as.SuffixParams) > 0 {
		result += "," + strings.Join(as.SuffixParams, ",")
	}

	result += ";"
	return result
}

// ExpressionStatement 代表一个表达式语句
type ExpressionStatement struct {
	Token      Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Parser 代表语法分析器
type Parser struct {
	l       *Lexer
	curTok  Token
	peekTok Token
	// drawingCounts 记录各画图函数已出现的次数，用于给「无变量名画图语句」生成唯一输出名。
	drawingCounts map[string]int
}

// NewParser 创建一个新的语法分析器
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, drawingCounts: map[string]int{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}
	program.Errors = []string{}

	for p.curTok.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// 语法校验：确保每行以分号结尾
		if p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
			errorMsg := fmt.Sprintf("第%d行第%d列：语法错误，语句必须以分号结尾，当前token: %s", p.curTok.Line, p.curTok.Column, p.curTok.Literal)
			Logger.Printf("第%d行第%d列：语法错误，语句必须以分号结尾，当前token: %s", p.curTok.Line, p.curTok.Column, p.curTok.Literal)
			program.Errors = append(program.Errors, errorMsg)
			// 返回错误程序，停止解析
			return program
		}

		p.nextToken()
	}
	Logger.Println("Parsed program with", len(program.Statements), "statements")
	for i, stmt := range program.Statements {
		if stmt != nil {
			str := ""
			if as, ok := stmt.(*AssignmentStatement); ok {
				if as.Name != nil {
					if as.IsOutputVar {
						str = as.Name.String() + " : "
					} else {
						str = as.Name.String() + " := "
					}
					if as.Value != nil {
						str += as.Value.String()
					} else {
						str += "<nil value>"
					}
					// 添加修饰符信息
					if len(as.SuffixParams) > 0 {
						str += "," + strings.Join(as.SuffixParams, ",")
					}
				} else {
					str = "<nil name>"
				}
			} else {
				str = stmt.String()
			}
			Logger.Printf("Statement %d: %s\n", i, str)
		} else {
			Logger.Printf("Statement %d: <nil>\n", i)
		}
	}
	return program
}

func (p *Parser) parseStatement() Statement {
	Logger.Println("Parsing statement, current token:", p.curTok.Literal)
	switch p.curTok.Type {
	case TokenIdentifier:
		if p.peekTok.Type == TokenColon || p.peekTok.Type == TokenColonEqual {
			Logger.Println("Found assignment statement")
			return p.parseAssignmentStatement()
		}
		// 如果不是赋值语句，尝试解析为表达式语句
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			// 带修饰符或画图函数调用 → 无变量名画图语句（如 DRAWTEXT_FIX(...),COLORRED; / DRAWICON(...);）
			if p.peekTok.Type == TokenComma || isDrawFuncCall(expr) {
				return p.parseDrawingStatement(expr)
			}
			// 消费分号（如果存在）
			if p.peekTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return &ExpressionStatement{Expression: expr}
		}
	default:
		// 尝试解析为表达式语句
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			// 带修饰符或画图函数调用 → 无变量名画图语句
			if p.peekTok.Type == TokenComma || isDrawFuncCall(expr) {
				return p.parseDrawingStatement(expr)
			}
			// 消费分号（如果存在）
			if p.peekTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return &ExpressionStatement{Expression: expr}
		}
	}
	return nil
}

// isDrawFuncCall 判断表达式是否为「已注册的画图函数」调用。
func isDrawFuncCall(expr Expression) bool {
	call, ok := expr.(*FunctionCall)
	if !ok {
		return false
	}
	id, ok := call.Function.(*Identifier)
	if !ok {
		return false
	}
	return IsDrawFunc(id.Value)
}

func (p *Parser) parseAssignmentStatement() *AssignmentStatement {
	stmt := &AssignmentStatement{Token: p.curTok, SuffixParams: []string{}}
	Logger.Println("Parsing assignment, current token:", p.curTok.Literal)

	// 已经检查过当前token是标识符
	stmt.Name = &Identifier{Token: p.curTok, Value: p.curTok.Literal}
	Logger.Println("Set name to:", stmt.Name.Value)

	// 检查赋值类型
	if p.peekTok.Type == TokenColonEqual {
		// 普通赋值 :=
		stmt.IsOutputVar = false
		Logger.Println("Found := assignment")
		p.nextToken() // 跳过 :=
	} else if p.peekTok.Type == TokenColon {
		// 画图赋值 :
		stmt.IsOutputVar = true
		Logger.Println("Found : assignment (drawing variable)")
		p.nextToken() // 跳过 :
	} else {
		Logger.Println("Expected colon or colon-equal, but not found")
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	p.parseSuffixParams(stmt)
	return stmt
}

// parseSuffixParams 解析赋值语句/画图语句末尾的修饰符（逗号分隔的标识符），并消费分号。
func (p *Parser) parseSuffixParams(stmt *AssignmentStatement) {
	if p.peekTok.Type == TokenComma {
		p.nextToken() // 跳过第一个逗号
		for p.peekTok.Type == TokenIdentifier {
			p.nextToken()
			stmt.SuffixParams = append(stmt.SuffixParams, p.curTok.Literal)
			Logger.Println("Added suffix param:", p.curTok.Literal)
			if p.peekTok.Type == TokenComma {
				p.nextToken() // 跳过逗号
			} else {
				break
			}
		}
	}
	// 有/无修饰符都消费分号
	if p.peekTok.Type == TokenSemicolon {
		p.nextToken()
	}
}

// parseDrawingStatement 解析「无变量名画图语句」，如 DRAWTEXT_FIX(...),COLORC08000;
// 规范化为 AssignmentStatement：变量名取函数名，IsOutputVar=true，Value 为函数调用表达式。
func (p *Parser) parseDrawingStatement(expr Expression) *AssignmentStatement {
	stmt := &AssignmentStatement{Token: p.curTok, SuffixParams: []string{}}
	stmt.IsOutputVar = true
	stmt.IsDrawingStatement = true
	stmt.Value = expr
	base := expr.String()
	if call, ok := expr.(*FunctionCall); ok {
		if id, ok := call.Function.(*Identifier); ok {
			base = id.Value
		}
	}
	// 多条同名画图语句：第 1 条用函数名，后续用「函数名#序号」，保证输出名唯一。
	p.drawingCounts[base]++
	name := base
	if n := p.drawingCounts[base]; n > 1 {
		name = fmt.Sprintf("%s#%d", base, n)
	}
	stmt.Name = &Identifier{Token: p.curTok, Value: name}
	p.parseSuffixParams(stmt)
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	Logger.Println("Parsing expression, current token:", p.curTok.Literal)
	prefix := p.parsePrefix(p.curTok.Type)
	if prefix == nil {
		Logger.Println("No prefix parser for token", p.curTok.Literal)
		return nil
	}
	leftExp := prefix()
	Logger.Println("Parsed prefix expression")

	// 检查是否是函数调用
	if p.peekTok.Type == TokenLParen {
		Logger.Println("Function call parsing, peek token:", p.peekTok.Literal)
		p.nextToken()
		leftExp = p.parseFunctionCall(leftExp)
		Logger.Println("Parsed function call, continuing with infix parsing")
		// parseExpressionList 已经跳过了右括号，所以当前 token 应该是右括号之后的 token
	}

	// 继续解析二元表达式，直到遇到分号、逗号、右括号或EOF，或者优先级不够
	// 逗号和右括号表示当前表达式的结束（在函数参数列表中等）
	for p.peekTok.Type != TokenSemicolon && p.peekTok.Type != TokenComma && p.peekTok.Type != TokenRParen && p.peekTok.Type != TokenEOF && precedence < p.peekPrecedence() {
		Logger.Println("Infix parsing, peek token:", p.peekTok.Literal)
		infix := p.parseInfix(leftExp, p.peekTok.Type)
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix()
	}

	return leftExp
}

func (p *Parser) parsePrefix(tokenType TokenType) func() Expression {
	switch tokenType {
	case TokenIdentifier:
		return p.parseIdentifier
	case TokenNumber:
		return p.parseNumberLiteral
	case TokenString:
		return p.parseStringLiteral
	case TokenLParen:
		return p.parseGroupedExpression
	case TokenMinus, TokenNot: // 支持负号作为一元运算符
		return p.parseUnaryExpression
	}
	return nil
}

func (p *Parser) parseInfix(left Expression, tokenType TokenType) func() Expression {
	switch tokenType {
	case TokenPlus, TokenMinus, TokenMultiply, TokenDivide, TokenAnd, TokenOr, TokenGreaterThan, TokenLessThan, TokenGreaterEqual, TokenLessEqual, TokenEqual, TokenNotEqual:
		return func() Expression { return p.parseBinaryExpression(left) }
	}
	return nil
}

func (p *Parser) parseBinaryExpression(left Expression) Expression {
	expression := &BinaryExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseUnaryExpression() Expression {
	expression := &UnaryExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
	}

	precedence := PREFIX
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseNumberLiteral() Expression {
	lit := &NumberLiteral{Token: p.curTok}
	value, err := strconv.ParseFloat(p.curTok.Literal, 64)
	if err != nil {
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	lit := &StringLiteral{Token: p.curTok}
	lit.Value = p.curTok.Literal
	return lit
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if exp == nil {
		return nil
	}
	if !p.expectPeek(TokenRParen) {
		return nil
	}
	return exp
}

func (p *Parser) parseFunctionCall(function Expression) Expression {
	exp := &FunctionCall{Token: p.curTok, Function: function}
	Logger.Println("Parsing function call for", function.String())
	exp.Arguments = p.parseExpressionList(TokenRParen)
	Logger.Println("Parsed", len(exp.Arguments), "arguments")
	return exp
}

func (p *Parser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}

	// 如果下一个token就是结束token，说明没有参数
	if p.peekTok.Type == end {
		p.nextToken()
		return list
	}

	// 跳过左括号
	p.nextToken()

	// 解析第一个参数
	if p.curTok.Type != end {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			list = append(list, expr)
		}
	}

	// 解析剩余参数
	for p.peekTok.Type == TokenComma {
		p.nextToken() // 跳过逗号
		p.nextToken() // 移动到下一个token
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			list = append(list, expr)
		}
	}

	// 确保消费结束token（右括号）
	// 先检查 p.curTok，如果是结束token就消费
	if p.curTok.Type == end {
		p.nextToken()
	} else if p.peekTok.Type == end {
		p.nextToken()
	}

	return list
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTok.Type == t {
		p.nextToken()
		return true
	}
	return false
}

const (
	_ int = iota
	LOWEST
	OR
	AND
	COMPARISON
	SUM
	PRODUCT
	PREFIX
	CALL
)

func (p *Parser) peekPrecedence() int {
	switch p.peekTok.Type {
	case TokenOr:
		return OR
	case TokenAnd:
		return AND
	case TokenGreaterThan, TokenLessThan, TokenGreaterEqual, TokenLessEqual, TokenEqual, TokenNotEqual:
		return COMPARISON
	case TokenPlus, TokenMinus:
		return SUM
	case TokenMultiply, TokenDivide:
		return PRODUCT
	case TokenLParen:
		return CALL
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	switch p.curTok.Type {
	case TokenOr:
		return OR
	case TokenAnd:
		return AND
	case TokenGreaterThan, TokenLessThan, TokenGreaterEqual, TokenLessEqual, TokenEqual, TokenNotEqual:
		return COMPARISON
	case TokenPlus, TokenMinus:
		return SUM
	case TokenMultiply, TokenDivide:
		return PRODUCT
	}
	return LOWEST
}
