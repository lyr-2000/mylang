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
	IsOutputVar bool // true表示是画图变量赋值(:), false表示普通赋值(:=)
	SuffixParams []string // 存储修饰符，如 COLORRED, NODRAW
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
}

// NewParser 创建一个新的语法分析器
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
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
			return &ExpressionStatement{Expression: expr}
		}
	default:
		// 尝试解析为表达式语句
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			return &ExpressionStatement{Expression: expr}
		}
	}
	return nil
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

	// 解析修饰符（逗号分隔的标识符）
	if p.peekTok.Type == TokenComma {
		p.nextToken() // 跳过第一个逗号
		for p.peekTok.Type == TokenIdentifier {
			p.nextToken()
			stmt.SuffixParams = append(stmt.SuffixParams, p.curTok.Literal)
			Logger.Println("Added suffix param:", p.curTok.Literal)
			
			// 如果下一个是逗号，继续解析
			if p.peekTok.Type == TokenComma {
				p.nextToken() // 跳过逗号
			} else {
				break
			}
		}
	}

	// 跳过分号（如果存在）
	if p.peekTok.Type == TokenSemicolon {
		p.nextToken()
	}

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
	}

	// 继续解析二元表达式，直到遇到分号、右括号或EOF，或者优先级不够
	for p.peekTok.Type != TokenSemicolon && p.peekTok.Type != TokenRParen && p.peekTok.Type != TokenEOF && precedence < p.peekPrecedence() {
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

	// 跳过右括号
	if p.peekTok.Type == end {
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
