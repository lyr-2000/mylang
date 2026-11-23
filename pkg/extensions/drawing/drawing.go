// Package drawing 定义画图指令的约定结构及内置的画图/文本函数，供宿主渲染。
//
// 画图函数（DRAWTEXT_FIX 等）返回 Command，引擎将其输出 Kind 记为 KindDraw、
// Raw 承载该结构、Values 为空，不再参与后续数值计算。
package drawing

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/mylang"
	"github.com/spf13/cast"
)

// drawFuncNames 画图函数名：其裸调用应被解析器识别为「无变量名画图语句」（即使不带修饰符）。
var drawFuncNames = []string{"DRAWTEXT_FIX", "DRAWTEXT", "DRAWICON", "DRAWNUMBER"}

func init() {
	for _, name := range drawFuncNames {
		mylang.RegisterDrawFunc(name)
	}
}

// Command 一条画图指令，宿主据此渲染。
type Command struct {
	// Func 画图函数名，如 "DRAWTEXT_FIX"。
	Func string
	// Cond 触发条件序列（0/1），nil 表示无条件全画。
	Cond []float64
	// X / Y 坐标（相对位置 0~1，固定坐标类画图用，DRAWTEXT_FIX）。
	X float64
	Y float64
	// Align 对齐方式（DRAWTEXT_FIX 的第 4 参数）。
	Align int
	// Text 文字内容（画文字类用）。
	Text string
	// Color 颜色，从 COLORxxxx 修饰符提取（如 "COLORC08000"）。
	Color string
	// Price 价格序列（按 K 线位置画图用，DRAWTEXT/DRAWICON/DRAWNUMBER 的第 2 参数）。
	Price []float64
	// Type 图标编号（DRAWICON 的第 3 参数）。
	Type int
	// Data 数字序列（DRAWNUMBER 的第 3 参数）。
	Data []float64
}

// Func 是画图/文本函数的统一签名，与 api.RegisterFunction 的包装函数一致。
type Func func(args []any) any

// Funcs 内置的画图/文本函数表（大写函数名 → 实现）。
var Funcs = map[string]Func{
	"DRAWTEXT_FIX": DrawTextFix,
	"DRAWTEXT":     DrawText,
	"DRAWICON":     DrawIcon,
	"DRAWNUMBER":   DrawNumber,
	"STRCAT":       StrCat,
	"CON2STR":      Con2Str,
	"VAR2STR":      Var2Str,
	"STRLEN":       StrLen,
	"SUBSTR":       SubStr,
	"STRCMP":       StrCmp,
	"FINDSTR":      FindStr,
}

// GetAllFuncNames 返回所有内置函数名。
func GetAllFuncNames() []string {
	names := make([]string, 0, len(Funcs))
	for name := range Funcs {
		names = append(names, name)
	}
	return names
}

// DrawTextFix 对应 DRAWTEXT_FIX(cond, x, y, align, text)：固定坐标画文字。
func DrawTextFix(args []any) any {
	if len(args) < 5 {
		return fmt.Errorf("DRAWTEXT_FIX 需要 5 个参数，实际 %d", len(args))
	}
	return Command{
		Func:  "DRAWTEXT_FIX",
		Cond:  toFloats(args[0]),
		X:     cast.ToFloat64(args[1]),
		Y:     cast.ToFloat64(args[2]),
		Align: cast.ToInt(args[3]),
		Text:  fmt.Sprint(args[4]),
	}
}

// DrawText 对应 DRAWTEXT(cond, price, text)：条件满足时在价格位置书写文字。
func DrawText(args []any) any {
	if len(args) < 3 {
		return fmt.Errorf("DRAWTEXT 需要 3 个参数，实际 %d", len(args))
	}
	return Command{
		Func:  "DRAWTEXT",
		Cond:  toFloats(args[0]),
		Price: toFloats(args[1]),
		Text:  fmt.Sprint(args[2]),
	}
}

// DrawIcon 对应 DRAWICON(cond, price, type)：条件满足时在价格位置画 type 号图标。
func DrawIcon(args []any) any {
	if len(args) < 3 {
		return fmt.Errorf("DRAWICON 需要 3 个参数，实际 %d", len(args))
	}
	return Command{
		Func:  "DRAWICON",
		Cond:  toFloats(args[0]),
		Price: toFloats(args[1]),
		Type:  cast.ToInt(args[2]),
	}
}

// DrawNumber 对应 DRAWNUMBER(cond, price, data)：条件满足时在价格位置书写数字。
func DrawNumber(args []any) any {
	if len(args) < 3 {
		return fmt.Errorf("DRAWNUMBER 需要 3 个参数，实际 %d", len(args))
	}
	return Command{
		Func:  "DRAWNUMBER",
		Cond:  toFloats(args[0]),
		Price: toFloats(args[1]),
		Data:  toFloats(args[2]),
	}
}

// StrCat 对应 STRCAT(a, b, ...)：把所有参数拼接为字符串。
func StrCat(args []any) any {
	s := ""
	for _, a := range args {
		s += fmt.Sprint(a)
	}
	return s
}

// Con2Str 对应 CON2STR(n, digits)：数字转字符串，digits 缺省为 2。
func Con2Str(args []any) any {
	if len(args) < 1 {
		return ""
	}
	digits := 2
	if len(args) >= 2 {
		digits = cast.ToInt(args[1])
	}
	return strconv.FormatFloat(cast.ToFloat64(args[0]), 'f', digits, 64)
}

// Var2Str 对应 VAR2STR(A, N)：把 A 的每个值转为字符串（小数位 N）。
// A 为序列时返回 []string，为标量时返回 string。
func Var2Str(args []any) any {
	if len(args) < 1 {
		return nil
	}
	digits := 2
	if len(args) >= 2 {
		digits = cast.ToInt(args[1])
	}
	switch x := args[0].(type) {
	case []float64:
		out := make([]string, len(x))
		for i, v := range x {
			out[i] = strconv.FormatFloat(v, 'f', digits, 64)
		}
		return out
	case indicators.Series:
		out := make([]string, len(x))
		for i, v := range x {
			out[i] = strconv.FormatFloat(float64(v), 'f', digits, 64)
		}
		return out
	}
	return strconv.FormatFloat(cast.ToFloat64(args[0]), 'f', digits, 64)
}

// StrLen 对应 STRLEN(S)：字符串长度（中文按 2 字节计，ASCII 按 1 字节，对齐 GBK 字节语义）。
func StrLen(args []any) any {
	if len(args) < 1 {
		return 0.0
	}
	n := 0
	for _, r := range fmt.Sprint(args[0]) {
		if r < 128 {
			n++
		} else {
			n += 2
		}
	}
	return float64(n)
}

// SubStr 对应 SUBSTR(S, A, N)：取 S 中从位置 A（1 开始）起 N 个字符。
func SubStr(args []any) any {
	if len(args) < 3 {
		return ""
	}
	runes := []rune(fmt.Sprint(args[0]))
	start := cast.ToInt(args[1]) - 1
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		return ""
	}
	end := start + cast.ToInt(args[2])
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// StrCmp 对应 STRCMP(A, B)：两字符串相等返回 1，否则 0。
func StrCmp(args []any) any {
	if len(args) < 2 {
		return 0.0
	}
	if fmt.Sprint(args[0]) == fmt.Sprint(args[1]) {
		return 1.0
	}
	return 0.0
}

// FindStr 对应 FINDSTR(A, B)：在 A 中查找到 B 返回 1，否则 0。
func FindStr(args []any) any {
	if len(args) < 2 {
		return 0.0
	}
	if strings.Contains(fmt.Sprint(args[0]), fmt.Sprint(args[1])) {
		return 1.0
	}
	return 0.0
}

// toFloats 把画图条件参数（[]float64 或 indicators.Series）转成 []float64。
func toFloats(v any) []float64 {
	switch x := v.(type) {
	case []float64:
		return x
	case indicators.Series:
		return []float64(x)
	}
	return nil
}
