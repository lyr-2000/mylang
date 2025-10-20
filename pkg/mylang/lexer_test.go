package mylang

import (
	"reflect"
	"testing"
)

func TestTrimComment(t *testing.T) {
	t.Run("trim-comment:1", func(t *testing.T) {
		code := "abc{comment}\ndef"
		got := TrimComment(code)
		want := "abc\ndef"
		if got != want {
			t.Errorf("TrimComment() = %v, want %v", got, want)
		}
	})
}

func TestLexerSuffixParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "单个修饰符",
			input: "test:HIGH>CLOSE,COLORRED;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "test"},
				{Type: TokenColon, Literal: ":"},
				{Type: TokenIdentifier, Literal: "HIGH"},
				{Type: TokenGreaterThan, Literal: ">"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "COLORRED"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
		{
			name:  "多个修饰符",
			input: "test:HIGH>CLOSE,COLORRED,NODRAW;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "test"},
				{Type: TokenColon, Literal: ":"},
				{Type: TokenIdentifier, Literal: "HIGH"},
				{Type: TokenGreaterThan, Literal: ">"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "COLORRED"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "NODRAW"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
		{
			name:  "普通赋值带修饰符",
			input: "test:=HIGH>CLOSE,COLORBLUE;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "test"},
				{Type: TokenColonEqual, Literal: ":="},
				{Type: TokenIdentifier, Literal: "HIGH"},
				{Type: TokenGreaterThan, Literal: ">"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "COLORBLUE"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
		{
			name:  "无修饰符",
			input: "test:HIGH>CLOSE;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "test"},
				{Type: TokenColon, Literal: ":"},
				{Type: TokenIdentifier, Literal: "HIGH"},
				{Type: TokenGreaterThan, Literal: ">"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			var tokens []Token
			for {
				token := lexer.NextToken()
				tokens = append(tokens, token)
				if token.Type == TokenEOF {
					break
				}
			}

			if !reflect.DeepEqual(tokens, tt.expected) {
				t.Errorf("Tokens = %v, want %v", tokens, tt.expected)
			}
		})
	}
}

func TestLexerSyntaxValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "有效语法-带分号",
			input:    "test:HIGH>CLOSE;",
			hasError: false,
		},
		{
			name:     "有效语法-带修饰符和分号",
			input:    "test:HIGH>CLOSE,COLORRED;",
			hasError: false,
		},
		{
			name:     "无效语法-缺少分号",
			input:    "test:HIGH>CLOSE",
			hasError: false, // 词法分析器不负责语法校验
		},
		{
			name:     "无效语法-多个语句缺少分号",
			input:    "test1:HIGH>CLOSE\ntest2:HIGH<CLOSE;",
			hasError: false, // 词法分析器不负责语法校验
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			var tokens []Token
			hasError := false
			
			for {
				token := lexer.NextToken()
				tokens = append(tokens, token)
				if token.Type == TokenError {
					hasError = true
				}
				if token.Type == TokenEOF {
					break
				}
			}

			if hasError != tt.hasError {
				t.Errorf("HasError = %v, want %v", hasError, tt.hasError)
			}
		})
	}
}

func TestLexerComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "复杂逻辑表达式",
			input: "result:HIGH>CLOSE AND LOW<CLOSE,COLORRED;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "result"},
				{Type: TokenColon, Literal: ":"},
				{Type: TokenIdentifier, Literal: "HIGH"},
				{Type: TokenGreaterThan, Literal: ">"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenAnd, Literal: "AND"},
				{Type: TokenIdentifier, Literal: "LOW"},
				{Type: TokenLessThan, Literal: "<"},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "COLORRED"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
		{
			name:  "函数调用表达式",
			input: "ma_result:MA(CLOSE,5),COLORBLUE;",
			expected: []Token{
				{Type: TokenIdentifier, Literal: "ma_result"},
				{Type: TokenColon, Literal: ":"},
				{Type: TokenIdentifier, Literal: "MA"},
				{Type: TokenLParen, Literal: "("},
				{Type: TokenIdentifier, Literal: "CLOSE"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenNumber, Literal: "5"},
				{Type: TokenRParen, Literal: ")"},
				{Type: TokenComma, Literal: ","},
				{Type: TokenIdentifier, Literal: "COLORBLUE"},
				{Type: TokenSemicolon, Literal: ";"},
				{Type: TokenEOF, Literal: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			var tokens []Token
			for {
				token := lexer.NextToken()
				tokens = append(tokens, token)
				if token.Type == TokenEOF {
					break
				}
			}

			if !reflect.DeepEqual(tokens, tt.expected) {
				t.Errorf("Tokens = %v, want %v", tokens, tt.expected)
			}
		})
	}
}