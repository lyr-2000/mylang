package mylang

// TokenType 定义了令牌的类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenNumber
	TokenIdentifier
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenLParen
	TokenRParen
	TokenSemicolon
	TokenColon
	TokenColonEqual
	TokenComma
	TokenAsterisk
	TokenEqual
	TokenAnd
	TokenOr
	TokenGreaterThan
	TokenLessThan
	TokenGreaterEqual
	TokenLessEqual
	TokenNotEqual
	TokenString
)

// Token 代表一个令牌
type Token struct {
	Type    TokenType
	Literal string
}

// Lexer 代表词法分析器
type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

// NewLexer 创建一个新的词法分析器
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '+':
		tok = Token{Type: TokenPlus, Literal: string(l.ch)}
		l.readChar()
	case '-':
		tok = Token{Type: TokenMinus, Literal: string(l.ch)}
		l.readChar()
	case '*':
		tok = Token{Type: TokenMultiply, Literal: string(l.ch)}
		l.readChar()
	case '/':
		tok = Token{Type: TokenDivide, Literal: string(l.ch)}
		l.readChar()
	case '(':
		tok = Token{Type: TokenLParen, Literal: string(l.ch)}
		l.readChar()
	case ')':
		tok = Token{Type: TokenRParen, Literal: string(l.ch)}
		l.readChar()
	case ';':
		tok = Token{Type: TokenSemicolon, Literal: string(l.ch)}
		l.readChar()
	case ':':
		if l.peekChar() == '=' {
			// 处理 := 赋值
			l.readChar()
			tok = Token{Type: TokenColonEqual, Literal: ":="}
		} else {
			// 处理单个 : 赋值（用于画图）
			tok = Token{Type: TokenColon, Literal: string(l.ch)}
		}
		l.readChar()
	case ',':
		tok = Token{Type: TokenComma, Literal: string(l.ch)}
		l.readChar()
	case '=':
		if l.peekChar() == '=' {
			// 处理 == 比较
			l.readChar()
			tok = Token{Type: TokenEqual, Literal: "=="}
		} else {
			tok = Token{Type: TokenEqual, Literal: string(l.ch)}
		}
		l.readChar()
	case '>':
		if l.peekChar() == '=' {
			// 处理 >= 比较
			l.readChar()
			tok = Token{Type: TokenGreaterEqual, Literal: ">="}
		} else {
			tok = Token{Type: TokenGreaterThan, Literal: string(l.ch)}
		}
		l.readChar()
	case '<':
		if l.peekChar() == '=' {
			// 处理 <= 比较
			l.readChar()
			tok = Token{Type: TokenLessEqual, Literal: "<="}
		} else if l.peekChar() == '>' {
			// 处理 != 比较
			l.readChar()
			tok = Token{Type: TokenNotEqual, Literal: "!="}
		} else {
			tok = Token{Type: TokenLessThan, Literal: string(l.ch)}
		}
		l.readChar()
	case '!':
		if l.peekChar() == '=' {
			// 处理 != 比较
			l.readChar()
			tok = Token{Type: TokenNotEqual, Literal: "!="}
		} else {
			tok = Token{Type: TokenError, Literal: string(l.ch)}
		}
		l.readChar()
	case '\'':
		tok.Type = TokenString
		tok.Literal = l.readString()
		return tok
	case 0:
		tok = Token{Type: TokenEOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = TokenIdentifier
			// 检查是否为关键字
			if tok.Literal == "AND" {
				tok.Type = TokenAnd
			} else if tok.Literal == "OR" || tok.Literal == "or" {
				tok.Type = TokenOr
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TokenNumber
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = Token{Type: TokenError, Literal: string(l.ch)}
		}
		l.readChar()
	}
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() string {
	pos := l.pos + 1 // 跳过开始的单引号
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	// 如果遇到结束的单引号，跳过它
	if l.ch == '\'' {
		l.readChar()
	}
	return l.input[pos:l.pos-1] // 去掉结束的单引号
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
