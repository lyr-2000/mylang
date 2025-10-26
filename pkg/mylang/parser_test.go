package mylang

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseAssignmentWithSuffixParams(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "单个修饰符",
			code:     "test:HIGH>CLOSE,COLORRED;",
			expected: []string{"COLORRED"},
		},
		{
			name:     "多个修饰符",
			code:     "test:HIGH>CLOSE,COLORRED,NODRAW;",
			expected: []string{"COLORRED", "NODRAW"},
		},
		{
			name:     "三个修饰符",
			code:     "test:HIGH>CLOSE,COLORRED,NODRAW,COLORBLUE;",
			expected: []string{"COLORRED", "NODRAW", "COLORBLUE"},
		},
		{
			name:     "无修饰符",
			code:     "test:HIGH>CLOSE;",
			expected: []string{},
		},
		{
			name:     "普通赋值带修饰符",
			code:     "test:=HIGH>CLOSE,COLORRED;",
			expected: []string{"COLORRED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.code)
			parser := NewParser(lexer)
			program := parser.ParseProgram()

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*AssignmentStatement)
			if !ok {
				t.Fatalf("Expected AssignmentStatement, got %T", program.Statements[0])
			}

			if !reflect.DeepEqual(stmt.SuffixParams, tt.expected) {
				t.Errorf("SuffixParams = %v, want %v", stmt.SuffixParams, tt.expected)
			}
		})
	}
}

func TestParseSyntaxValidation(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		shouldError  bool
		expectedLine int // 期望的错误行号
	}{
		{
			name:         "有效语法-带分号",
			code:         "test:HIGH>CLOSE;",
			shouldError:  false,
			expectedLine: 0,
		},
		{
			name:         "有效语法-带修饰符和分号",
			code:         "test:HIGH>CLOSE,COLORRED;",
			shouldError:  false,
			expectedLine: 0,
		},
		{
			name:         "无效语法-缺少分号",
			code:         "test:HIGH>CLOSE",
			shouldError:  true,
			expectedLine: 1,
		},
		{
			name:         "无效语法-多个语句缺少分号",
			code:         "test1:HIGH>CLOSE\ntest2:HIGH<CLOSE;",
			shouldError:  true,
			expectedLine: 1,
		},
		{
			name:         "多行代码错误",
			code:         "line1:HIGH>CLOSE;\nline2:HIGH<CLOSE\nline3:HIGH=CLOSE;",
			shouldError:  true,
			expectedLine: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.code)
			parser := NewParser(lexer)
			program := parser.ParseProgram()

			if tt.shouldError {
				// 对于应该出错的语法，期望有错误
				if len(program.Errors) == 0 {
					t.Errorf("Expected syntax error for code: %s, but got none", tt.code)
				} else {
					// 检查错误信息是否包含行号
					errorMsg := program.Errors[0]
					if tt.expectedLine > 0 {
						expectedLineStr := fmt.Sprintf("第%d行", tt.expectedLine)
						if !strings.Contains(errorMsg, expectedLineStr) {
							t.Errorf("Expected error message to contain line number %d, but got: %s", tt.expectedLine, errorMsg)
						}
					}
				}
			} else {
				// 对于有效语法，期望能正常解析且无错误
				if len(program.Errors) > 0 {
					t.Errorf("Expected no syntax errors, but got: %v", program.Errors)
				}
				if len(program.Statements) == 0 {
					t.Errorf("Expected statements to be parsed, but got none")
				}
			}
		})
	}
}

func TestParseComplexExpressionsWithSuffixParams(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedName   string
		expectedParams []string
		isDrawingVar   bool
	}{
		{
			name:           "复杂表达式带修饰符",
			code:           "result:HIGH>CLOSE AND LOW<OPEN,COLORRED,NODRAW;",
			expectedName:   "result",
			expectedParams: []string{"COLORRED", "NODRAW"},
			isDrawingVar:   true,
		},
		{
			name:           "函数调用带修饰符",
			code:           "func_result:MA(CLOSE,5),COLORBLUE;",
			expectedName:   "func_result",
			expectedParams: []string{"COLORBLUE"},
			isDrawingVar:   true,
		},
		{
			name:           "普通赋值带修饰符",
			code:           "normal_var:=HIGH+CLOSE,COLORGREEN;",
			expectedName:   "normal_var",
			expectedParams: []string{"COLORGREEN"},
			isDrawingVar:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.code)
			parser := NewParser(lexer)
			program := parser.ParseProgram()

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*AssignmentStatement)
			if !ok {
				t.Fatalf("Expected AssignmentStatement, got %T", program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("Name = %v, want %v", stmt.Name.Value, tt.expectedName)
			}

			if !reflect.DeepEqual(stmt.SuffixParams, tt.expectedParams) {
				t.Errorf("SuffixParams = %v, want %v", stmt.SuffixParams, tt.expectedParams)
			}

			if stmt.IsOutputVar != tt.isDrawingVar {
				t.Errorf("IsDrawingVar = %v, want %v", stmt.IsOutputVar, tt.isDrawingVar)
			}
		})
	}
}

func TestAssignmentStatementString(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *AssignmentStatement
		expected string
	}{
		{
			name: "画图变量带修饰符",
			stmt: &AssignmentStatement{
				Name:         &Identifier{Value: "test"},
				Value:        &Identifier{Value: "HIGH"},
				IsOutputVar: true,
				SuffixParams: []string{"COLORRED", "NODRAW"},
			},
			expected: "test : HIGH,COLORRED,NODRAW;",
		},
		{
			name: "普通变量带修饰符",
			stmt: &AssignmentStatement{
				Name:         &Identifier{Value: "test"},
				Value:        &Identifier{Value: "HIGH"},
				IsOutputVar: false,
				SuffixParams: []string{"COLORBLUE"},
			},
			expected: "test := HIGH,COLORBLUE;",
		},
		{
			name: "无修饰符",
			stmt: &AssignmentStatement{
				Name:         &Identifier{Value: "test"},
				Value:        &Identifier{Value: "HIGH"},
				IsOutputVar: true,
				SuffixParams: []string{},
			},
			expected: "test : HIGH;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stmt.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
