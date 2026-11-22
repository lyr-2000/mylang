// Package fastcall 提供指标函数的「去反射」类型化调用器。
//
// 指标函数本来有静态签名（如 func(Series,int) Series），但注册到解释器时被包成
// 统一的 func([]any) any，运行时用 reflect.Call 恢复签名调用，带来装箱 + 反射开销。
// 本包在「注册时」用类型断言生成类型化闭包，运行时零反射直接调用。
//
// 不改父包 indicators：指标函数都是导出的，直接引用函数值做类型断言。
// 兼容性：未列入覆盖表或签名不匹配时 Invoker 返回 ok=false，调用方兜底走反射。
//
// 用户自定义函数也可通过 Register 注册具体签名函数，获得同样的去反射能力。
package fastcall

import (
	"fmt"
	"sync"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
)

// fastFuncs 是「去反射」覆盖的函数表（函数名 → 导出函数值）。
// mu 保护该表：Register 可能由用户并发调用，Invoker 只读。
var (
	fastFuncs = map[string]any{
		// 一元 Series → Series
		"ABS": indicators.ABS, "LN": indicators.LN, "SQRT": indicators.SQRT,
		"SIN": indicators.SIN, "COS": indicators.COS, "TAN": indicators.TAN,
		"NZ": indicators.NZ, "NUMBER": indicators.Number,
		"TOPRANGE": indicators.TOPRANGE, "LOWRANGE": indicators.LOWRANGE,
		"QRR": indicators.QRR, "CONST": indicators.GetConst,
		// Series,int → Series
		"DIFF": indicators.DIFF, "STD": indicators.STD, "SUM": indicators.SUM,
		"MA": indicators.MA, "EMA": indicators.EMA, "WMA": indicators.WMA,
		"HHV": indicators.HHV, "LLV": indicators.LLV, "RSI": indicators.RSI,
		"AVEDEV": indicators.AVEDEV, "SLOPE": indicators.SLOPE, "FORCAST": indicators.FORCAST,
		"DSMA": indicators.DSMA, "HHVBARS": indicators.HHVBARS, "LLVBARS": indicators.LLVBARS,
		// Series,float64 → Series
		"POW": indicators.POW, "MAXS": indicators.MAXS, "DMA": indicators.DMA,
		// Series,Series → Series
		"MAX": indicators.MAX, "MIN": indicators.MIN, "OBV": indicators.OBV,
		"ADD": indicators.ADD, "SUB": indicators.SUB, "MUL": indicators.MUL, "DIV": indicators.DIV,
		// Series,any → Series
		"REF": indicators.REF, "SUMBARS": indicators.SUMBARS,
		// Series,int,float64 → Series
		"SMA": indicators.SMA,
		// []bool → Series
		"BARSLAST": indicators.BARSLAST, "BARSLASTCOUNT": indicators.BARSLASTCOUNT,
		// []bool,int → Series / []bool
		"COUNT": indicators.COUNT, "EVERY": indicators.EVERY, "EXIST": indicators.EXIST,
		"BARSSINCE": indicators.BARSSINCE, "BARSSINCEN": indicators.BARSSINCEN,
		"FILTER": indicators.FILTER,
		// Series,Series → []bool
		"CROSS":       indicators.CROSS,
		"GreaterThan": indicators.GreaterThan, "LessThan": indicators.LessThan,
		"LessThanOrEqual": indicators.LessThanOrEqual, "Equal": indicators.Equal,
		"NotEqual": indicators.NotEqual, "GreaterThanOrEqual": indicators.GreaterThanOrEqual,
		// []bool,Series,Series → Series
		"IF": indicators.IF,
		// 单返回值多参数（多输入 OHLCV）
		"CCI": indicators.CCI, "ATR": indicators.ATR, "CR": indicators.CR,
		"VR": indicators.VR, "BBI": indicators.BBI, "SAR": indicators.SAR,
		// 标量
		"RD": indicators.RD, "RET": indicators.RET,
		// []bool,Series → Series
		"VALUEWHEN": indicators.VALUEWHEN,
		// Series,Series,Series → []bool / Series,Series,int → []bool
		"BETWEEN": indicators.BETWEEN, "LONGCROSS": indicators.LONGCROSS,
	}
	mu sync.RWMutex
)

// Invoker 返回 name 对应的类型化调用器；未覆盖或签名不匹配返回 ok=false（调用方兜底反射）。
func Invoker(name string) (func([]any) (any, error), bool) {
	mu.RLock()
	fn, ok := fastFuncs[name]
	mu.RUnlock()
	if !ok {
		return nil, false
	}
	inv := makeTypedInvoker(fn)
	if inv == nil {
		return nil, false
	}
	return inv, true
}

// Register 注册一个「具体签名」的自定义函数到去反射表，返回可直接传给
// executor.RegisterFunction 的包装函数。签名必须是 makeTypedInvoker 支持的
// 类型之一，否则返回 ok=false；name 已存在（含内置函数）时拒绝覆盖。
func Register(name string, fn any) (wrap func([]any) any, ok bool) {
	inv := makeTypedInvoker(fn)
	if inv == nil {
		return nil, false
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := fastFuncs[name]; exists {
		return nil, false
	}
	fastFuncs[name] = fn
	return func(args []any) any {
		b, err := inv(args)
		if err != nil {
			return err // executor 的 safeCall 会识别 error 返回值
		}
		return b
	}, true
}

// makeTypedInvoker 按函数静态签名生成「类型化调用器」闭包；签名不匹配返回 nil。
func makeTypedInvoker(fn any) func([]any) (any, error) {
	switch f := fn.(type) {
	// ---- 单返回值 ----
	case func(indicators.Series) indicators.Series:
		return func(args []any) (any, error) {
			s, err := oneSeries(args)
			if err != nil {
				return nil, err
			}
			return f(s), nil
		}
	case func(indicators.Series, int) indicators.Series:
		return func(args []any) (any, error) {
			s, n, err := seriesInt(args)
			if err != nil {
				return nil, err
			}
			return f(s, n), nil
		}
	case func(indicators.Series, float64) indicators.Series:
		return func(args []any) (any, error) {
			s, n, err := seriesFloat(args)
			if err != nil {
				return nil, err
			}
			return f(s, n), nil
		}
	case func(indicators.Series, indicators.Series) indicators.Series:
		return func(args []any) (any, error) {
			a, b, err := twoSeries(args)
			if err != nil {
				return nil, err
			}
			return f(a, b), nil
		}
	case func(indicators.Series, any) indicators.Series:
		return func(args []any) (any, error) {
			s, _, err := seriesAny(args)
			if err != nil {
				return nil, err
			}
			return f(s, args[1]), nil
		}
	case func(indicators.Series, int, float64) indicators.Series:
		return func(args []any) (any, error) {
			s, n, m, err := seriesIntFloat(args)
			if err != nil {
				return nil, err
			}
			return f(s, n, m), nil
		}
	case func([]bool) indicators.Series:
		return func(args []any) (any, error) {
			b, err := oneBool(args)
			if err != nil {
				return nil, err
			}
			return f(b), nil
		}
	case func([]bool, int) indicators.Series:
		return func(args []any) (any, error) {
			b, n, err := boolInt(args)
			if err != nil {
				return nil, err
			}
			return f(b, n), nil
		}
	case func([]bool, int) []bool:
		return func(args []any) (any, error) {
			b, n, err := boolInt(args)
			if err != nil {
				return nil, err
			}
			return f(b, n), nil
		}
	case func(indicators.Series, indicators.Series) []bool:
		return func(args []any) (any, error) {
			a, b, err := twoSeries(args)
			if err != nil {
				return nil, err
			}
			return f(a, b), nil
		}
	case func([]bool, indicators.Series, indicators.Series) indicators.Series:
		return func(args []any) (any, error) {
			c, a, b, err := boolTwoSeries(args)
			if err != nil {
				return nil, err
			}
			return f(c, a, b), nil
		}
	case func(indicators.Series, indicators.Series, indicators.Series, int) indicators.Series:
		// CCI / ATR / CR
		return func(args []any) (any, error) {
			a, b, c, n, err := threeSeriesInt(args)
			if err != nil {
				return nil, err
			}
			return f(a, b, c, n), nil
		}
	case func(indicators.Series, indicators.Series, int) indicators.Series:
		// VR
		return func(args []any) (any, error) {
			a, b, n, err := twoSeriesInt(args)
			if err != nil {
				return nil, err
			}
			return f(a, b, n), nil
		}
	case func(indicators.Series, int, int, int, int) indicators.Series:
		// BBI
		return func(args []any) (any, error) {
			s, a, b, c, d, err := seriesFourInt(args)
			if err != nil {
				return nil, err
			}
			return f(s, a, b, c, d), nil
		}
	case func(indicators.Series, indicators.Series, int, int, int) indicators.Series:
		// SAR
		return func(args []any) (any, error) {
			a, b, n, s, m, err := twoSeriesThreeInt(args)
			if err != nil {
				return nil, err
			}
			return f(a, b, n, s, m), nil
		}
	case func([]bool, indicators.Series) indicators.Series:
		// VALUEWHEN
		return func(args []any) (any, error) {
			b, s, err := boolSeries(args)
			if err != nil {
				return nil, err
			}
			return f(b, s), nil
		}
	case func(indicators.Series, indicators.Series, indicators.Series) []bool:
		// BETWEEN
		return func(args []any) (any, error) {
			a, b, c, err := threeSeries(args)
			if err != nil {
				return nil, err
			}
			return f(a, b, c), nil
		}
	case func(indicators.Series, indicators.Series, int) []bool:
		// LONGCROSS
		return func(args []any) (any, error) {
			a, b, n, err := twoSeriesInt(args)
			if err != nil {
				return nil, err
			}
			return f(a, b, n), nil
		}
	case func(float64, int) float64:
		// RD
		return func(args []any) (any, error) {
			if len(args) != 2 {
				return nil, argCountErr(2, len(args))
			}
			n, ok1 := asFloat(args[0])
			d, ok2 := asInt(args[1])
			if !ok1 || !ok2 {
				return nil, argConvErr(0, "float64,int")
			}
			return f(n, d), nil
		}
	case func(indicators.Series, int) float64:
		// RET
		return func(args []any) (any, error) {
			s, n, err := seriesInt(args)
			if err != nil {
				return nil, err
			}
			return f(s, n), nil
		}

	}
	return nil
}

// ---- 参数解构 helper（减少重复）----

func oneSeries(args []any) (indicators.Series, error) {
	if len(args) != 1 {
		return nil, argCountErr(1, len(args))
	}
	s, ok := asSeries(args[0], args...)
	if !ok {
		return nil, argConvErr(0, "Series")
	}
	return s, nil
}

func seriesInt(args []any) (indicators.Series, int, error) {
	if len(args) != 2 {
		return nil, 0, argCountErr(2, len(args))
	}
	s, ok1 := asSeries(args[0], args...)
	n, ok2 := asInt(args[1])
	if !ok1 || !ok2 {
		return nil, 0, argConvErr(0, "Series,int")
	}
	return s, n, nil
}

func seriesFloat(args []any) (indicators.Series, float64, error) {
	if len(args) != 2 {
		return nil, 0, argCountErr(2, len(args))
	}
	s, ok1 := asSeries(args[0], args...)
	n, ok2 := asFloat(args[1])
	if !ok1 || !ok2 {
		return nil, 0, argConvErr(0, "Series,float64")
	}
	return s, n, nil
}

func twoSeries(args []any) (indicators.Series, indicators.Series, error) {
	if len(args) != 2 {
		return nil, nil, argCountErr(2, len(args))
	}
	a, ok1 := asSeries(args[0], args...)
	b, ok2 := asSeries(args[1], args...)
	if !ok1 || !ok2 {
		return nil, nil, argConvErr(0, "Series,Series")
	}
	return a, b, nil
}

func threeSeries(args []any) (indicators.Series, indicators.Series, indicators.Series, error) {
	if len(args) != 3 {
		return nil, nil, nil, argCountErr(3, len(args))
	}
	a, ok1 := asSeries(args[0], args...)
	b, ok2 := asSeries(args[1], args...)
	c, ok3 := asSeries(args[2], args...)
	if !ok1 || !ok2 || !ok3 {
		return nil, nil, nil, argConvErr(0, "Series,Series,Series")
	}
	return a, b, c, nil
}

func seriesAny(args []any) (indicators.Series, any, error) {
	if len(args) != 2 {
		return nil, nil, argCountErr(2, len(args))
	}
	s, ok := asSeries(args[0], args...)
	if !ok {
		return nil, nil, argConvErr(0, "Series")
	}
	return s, args[1], nil
}

func seriesIntFloat(args []any) (indicators.Series, int, float64, error) {
	if len(args) != 3 {
		return nil, 0, 0, argCountErr(3, len(args))
	}
	s, ok1 := asSeries(args[0], args...)
	n, ok2 := asInt(args[1])
	m, ok3 := asFloat(args[2])
	if !ok1 || !ok2 || !ok3 {
		return nil, 0, 0, argConvErr(0, "Series,int,float64")
	}
	return s, n, m, nil
}

func oneBool(args []any) ([]bool, error) {
	if len(args) != 1 {
		return nil, argCountErr(1, len(args))
	}
	b, ok := asBoolSlice(args[0], args...)
	if !ok {
		return nil, argConvErr(0, "[]bool")
	}
	return b, nil
}

func boolInt(args []any) ([]bool, int, error) {
	if len(args) != 2 {
		return nil, 0, argCountErr(2, len(args))
	}
	b, ok1 := asBoolSlice(args[0], args...)
	n, ok2 := asInt(args[1])
	if !ok1 || !ok2 {
		return nil, 0, argConvErr(0, "[]bool,int")
	}
	return b, n, nil
}

func boolSeries(args []any) ([]bool, indicators.Series, error) {
	if len(args) != 2 {
		return nil, nil, argCountErr(2, len(args))
	}
	b, ok1 := asBoolSlice(args[0], args...)
	s, ok2 := asSeries(args[1], args...)
	if !ok1 || !ok2 {
		return nil, nil, argConvErr(0, "[]bool,Series")
	}
	return b, s, nil
}

func boolTwoSeries(args []any) ([]bool, indicators.Series, indicators.Series, error) {
	if len(args) != 3 {
		return nil, nil, nil, argCountErr(3, len(args))
	}
	c, ok1 := asBoolSlice(args[0], args...)
	a, ok2 := asSeries(args[1], args...)
	b, ok3 := asSeries(args[2], args...)
	if !ok1 || !ok2 || !ok3 {
		return nil, nil, nil, argConvErr(0, "[]bool,Series,Series")
	}
	return c, a, b, nil
}

func seriesFourInt(args []any) (indicators.Series, int, int, int, int, error) {
	if len(args) != 5 {
		return nil, 0, 0, 0, 0, argCountErr(5, len(args))
	}
	s, ok1 := asSeries(args[0], args...)
	a, ok2 := asInt(args[1])
	b, ok3 := asInt(args[2])
	c, ok4 := asInt(args[3])
	d, ok5 := asInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return nil, 0, 0, 0, 0, argConvErr(0, "Series,int,int,int,int")
	}
	return s, a, b, c, d, nil
}

func twoSeriesInt(args []any) (indicators.Series, indicators.Series, int, error) {
	if len(args) != 3 {
		return nil, nil, 0, argCountErr(3, len(args))
	}
	a, ok1 := asSeries(args[0], args...)
	b, ok2 := asSeries(args[1], args...)
	n, ok3 := asInt(args[2])
	if !ok1 || !ok2 || !ok3 {
		return nil, nil, 0, argConvErr(0, "Series,Series,int")
	}
	return a, b, n, nil
}

func twoSeriesThreeInt(args []any) (indicators.Series, indicators.Series, int, int, int, error) {
	if len(args) != 5 {
		return nil, nil, 0, 0, 0, argCountErr(5, len(args))
	}
	a, ok1 := asSeries(args[0], args...)
	b, ok2 := asSeries(args[1], args...)
	n, ok3 := asInt(args[2])
	s, ok4 := asInt(args[3])
	m, ok5 := asInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return nil, nil, 0, 0, 0, argConvErr(0, "Series,Series,int,int,int")
	}
	return a, b, n, s, m, nil
}

func threeSeriesInt(args []any) (indicators.Series, indicators.Series, indicators.Series, int, error) {
	if len(args) != 4 {
		return nil, nil, nil, 0, argCountErr(4, len(args))
	}
	a, ok1 := asSeries(args[0], args...)
	b, ok2 := asSeries(args[1], args...)
	c, ok3 := asSeries(args[2], args...)
	n, ok4 := asInt(args[3])
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, nil, nil, 0, argConvErr(0, "Series,Series,Series,int")
	}
	return a, b, c, n, nil
}

// ---- 参数转换（类型断言，零反射）----

// asSeries 把参数转成 Series（零拷贝：[]float64 与 Series 共享底层数组）。
// 复刻 convertType 的语义：标量（float64/int）广播到目标长度，[]bool → []float64。
func asSeries(a any, args ...any) (indicators.Series, bool) {
	switch v := a.(type) {
	case indicators.Series:
		return v, true
	case []float64:
		return indicators.Series(v), true
	case []bool:
		out := make(indicators.Series, len(v))
		for i, b := range v {
			if b {
				out[i] = 1
			}
		}
		return out, true
	case float64:
		n := maxSliceLen(args)
		out := make(indicators.Series, n)
		for i := range out {
			out[i] = v
		}
		return out, true
	case int:
		n := maxSliceLen(args)
		out := make(indicators.Series, n)
		for i := range out {
			out[i] = float64(v)
		}
		return out, true
	}
	return nil, false
}

// maxSliceLen 返回参数中最长 slice 的长度（用于标量广播），与 getMaxSliceLength 一致。
func maxSliceLen(args []any) int {
	n := 0
	for _, a := range args {
		switch v := a.(type) {
		case []float64:
			if len(v) > n {
				n = len(v)
			}
		case indicators.Series:
			if len(v) > n {
				n = len(v)
			}
		case []bool:
			if len(v) > n {
				n = len(v)
			}
		}
	}
	return n
}

func asInt(a any) (int, bool) {
	switch v := a.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	}
	return 0, false
}

func asFloat(a any) (float64, bool) {
	switch v := a.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	}
	return 0, false
}

// asBoolSlice 把参数转成 []bool。比较运算的结果是 []float64（0/1），
// 与 convertType 语义一致：非 0 → true；bool 标量广播到目标长度。
func asBoolSlice(a any, args ...any) ([]bool, bool) {
	switch v := a.(type) {
	case []bool:
		return v, true
	case indicators.Series:
		out := make([]bool, len(v))
		for i, x := range v {
			out[i] = x != 0.0
		}
		return out, true
	case []float64:
		out := make([]bool, len(v))
		for i, x := range v {
			out[i] = x != 0.0
		}
		return out, true
	case bool:
		n := maxSliceLen(args)
		out := make([]bool, n)
		for i := range out {
			out[i] = v
		}
		return out, true
	}
	return nil, false
}

func argCountErr(want, got int) error {
	return fmt.Errorf("argument count mismatch: expected %d, got %d", want, got)
}

func argConvErr(idx int, want string) error {
	return fmt.Errorf("cannot convert argument %d to %s", idx, want)
}
