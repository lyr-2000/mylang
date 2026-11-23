package api

import (
	"testing"

	"github.com/lyr-2000/mylang/pkg/extensions/drawing"
)

// registerDataFuncs 注册「数据源函数」（板块/财务/K线位置）。
// 这些依赖宿主外部数据，不在 drawing 子包内置；STRCAT/CON2STR/DRAWTEXT_FIX 已内置。
func registerDataFuncs(ex *MaiExecutor) {
	ex.RegisterFunction("HYBLOCK", func(args []interface{}) interface{} { return "医药" })
	ex.RegisterFunction("GNBLOCK", func(args []interface{}) interface{} { return "生物制药" })
	ex.RegisterFunction("DYBLOCK", func(args []interface{}) interface{} { return "上海" })
	ex.RegisterFunction("FINANCE", func(args []interface{}) interface{} {
		return 3.9e9 // 伪财务数据（总市值/流通市值）
	})
	ex.RegisterFunction("ISLASTBAR", func(args []interface{}) interface{} {
		return []float64{0, 0, 0, 0, 1} // 最后一根为 1
	})
}

const blockFormula = `
行业信息:=STRCAT('行业: ',STRCAT('',HYBLOCK));
概念信息:=STRCAT(' 概念: ',GNBLOCK);
地域信息:=STRCAT(' 地域: ',DYBLOCK);

DRAWTEXT_FIX(ISLASTBAR,0.00,0.01,0,行业信息),COLORC08000;
DRAWTEXT_FIX(ISLASTBAR,0.00,0.09,0,地域信息),COLORC08000;
DRAWTEXT_FIX(ISLASTBAR AND DYBLOCK,0.00,0.17,0,概念信息),COLORC08000;
DRAWTEXT_FIX(ISLASTBAR AND DYBLOCK,0.00,0.24,0,STRCAT('总市值: ', CON2STR( FINANCE(39)/100000000 ,2)  )),COLORC08000;
DRAWTEXT_FIX(ISLASTBAR AND DYBLOCK,0.00,0.32,0,STRCAT('流通市值: ', CON2STR( FINANCE(40)/100000000 ,2)  )),COLORC08000;
`

// TestBlockFormulaCompat 注册伪函数后，验证板块/财务/画图公式语法兼容（V1 与 V2）。
func TestBlockFormulaCompat(t *testing.T) {
	for name, ex := range map[string]*MaiExecutor{
		"v1": NewMaiExecutorV1(),
		"v2": NewMaiExecutor(),
	} {
		registerDataFuncs(ex)
		outs, err := ex.Run(blockFormula, nil)
		if err != nil {
			t.Fatalf("[%s] 运行失败: %v", name, err)
		}
		if len(outs) != 5 {
			t.Fatalf("[%s] 输出数 = %d, want 5", name, len(outs))
		}
		for i, o := range outs {
			if o.Kind != KindDraw {
				t.Fatalf("[%s] 输出 %d Kind=%d, want KindDraw", name, i, o.Kind)
			}
			dc, ok := o.Raw.(drawing.Command)
			if !ok {
				t.Fatalf("[%s] 输出 %d Raw 类型 %T, want drawing.Command", name, i, o.Raw)
			}
			if dc.Color != "COLORC08000" {
				t.Fatalf("[%s] 输出 %d Color=%q, want COLORC08000", name, i, dc.Color)
			}
			t.Logf("[%s] %s func=%s color=%s text=%q x=%v y=%v align=%d",
				name, o.Name, dc.Func, dc.Color, dc.Text, dc.X, dc.Y, dc.Align)
		}
	}
}

// TestStringFuncs 验证内置字符串函数（V1/V2 一致）。
func TestStringFuncs(t *testing.T) {
	for name, ex := range map[string]*MaiExecutor{
		"v1": NewMaiExecutorV1(),
		"v2": NewMaiExecutor(),
	} {
		// STRLEN('通达信') = 6（GBK 字节语义：3 汉字 × 2）
		outs, err := ex.Run("R:STRLEN('通达信');", nil)
		if err != nil {
			t.Fatalf("[%s] STRLEN: %v", name, err)
		}
		if got := outs[0].Raw; got != float64(6) {
			t.Fatalf("[%s] STRLEN('通达信') = %v, want 6", name, got)
		}

		// SUBSTR('通达信',2,2) = '达信'
		outs, err = ex.Run("R:SUBSTR('通达信',2,2);", nil)
		if err != nil {
			t.Fatalf("[%s] SUBSTR: %v", name, err)
		}
		if got := outs[0].Raw; got != "达信" {
			t.Fatalf("[%s] SUBSTR('通达信',2,2) = %v, want 达信", name, got)
		}

		// STRCMP('软件服务','软件服务') = 1
		outs, err = ex.Run("R:STRCMP('软件服务','软件服务');", nil)
		if err != nil {
			t.Fatalf("[%s] STRCMP: %v", name, err)
		}
		if got := outs[0].Raw; got != float64(1) {
			t.Fatalf("[%s] STRCMP(相等) = %v, want 1", name, got)
		}

		// FINDSTR('多头开仓','开仓') = 1
		outs, err = ex.Run("R:FINDSTR('多头开仓','开仓');", nil)
		if err != nil {
			t.Fatalf("[%s] FINDSTR: %v", name, err)
		}
		if got := outs[0].Raw; got != float64(1) {
			t.Fatalf("[%s] FINDSTR = %v, want 1", name, got)
		}

		// VAR2STR(C,3)：序列转字符串序列
		outs, err = ex.Run("R:VAR2STR(C,3);", map[string][]float64{"C": {1.23456, 2.34567}})
		if err != nil {
			t.Fatalf("[%s] VAR2STR: %v", name, err)
		}
		ss, ok := outs[0].Raw.([]string)
		if !ok {
			t.Fatalf("[%s] VAR2STR Raw 类型 %T, want []string", name, outs[0].Raw)
		}
		if len(ss) != 2 || ss[0] != "1.235" || ss[1] != "2.346" {
			t.Fatalf("[%s] VAR2STR(C,3) = %v, want [1.235 2.346]", name, ss)
		}
	}
}

// TestDrawTextIcon 验证 DRAWTEXT/DRAWICON 按价格位置画图（不带修饰符也识别为画图语句）。
func TestDrawTextIcon(t *testing.T) {
	for name, ex := range map[string]*MaiExecutor{
		"v1": NewMaiExecutorV1(),
		"v2": NewMaiExecutor(),
	} {
		// DRAWTEXT(C>10,C,'买')：不带修饰符
		outs, err := ex.Run("DRAWTEXT(C>10,C,'买');", map[string][]float64{"C": {5, 15}})
		if err != nil {
			t.Fatalf("[%s] DRAWTEXT: %v", name, err)
		}
		if len(outs) != 1 {
			t.Fatalf("[%s] DRAWTEXT 输出数 = %d, want 1", name, len(outs))
		}
		dc, ok := outs[0].Raw.(drawing.Command)
		if !ok {
			t.Fatalf("[%s] DRAWTEXT Raw 类型 %T", name, outs[0].Raw)
		}
		if dc.Func != "DRAWTEXT" || dc.Text != "买" {
			t.Fatalf("[%s] DRAWTEXT Command = %+v", name, dc)
		}
		if len(dc.Cond) != 2 || dc.Cond[0] != 0 || dc.Cond[1] != 1 {
			t.Fatalf("[%s] DRAWTEXT Cond = %v, want [0 1]", name, dc.Cond)
		}
		if len(dc.Price) != 2 || dc.Price[0] != 5 || dc.Price[1] != 15 {
			t.Fatalf("[%s] DRAWTEXT Price = %v, want [5 15]", name, dc.Price)
		}

		// DRAWICON(C>10,C,1)：不带修饰符
		outs, err = ex.Run("DRAWICON(C>10,C,1);", map[string][]float64{"C": {5, 15}})
		if err != nil {
			t.Fatalf("[%s] DRAWICON: %v", name, err)
		}
		dc, ok = outs[0].Raw.(drawing.Command)
		if !ok {
			t.Fatalf("[%s] DRAWICON Raw 类型 %T", name, outs[0].Raw)
		}
		if dc.Func != "DRAWICON" || dc.Type != 1 {
			t.Fatalf("[%s] DRAWICON Command = %+v", name, dc)
		}

		// DRAWICON 带 DRAWABOVE 修饰符
		outs, err = ex.Run("DRAWICON(C>10,C,1),DRAWABOVE;", map[string][]float64{"C": {5, 15}})
		if err != nil {
			t.Fatalf("[%s] DRAWICON+DRAWABOVE: %v", name, err)
		}
		dc, ok = outs[0].Raw.(drawing.Command)
		if !ok || dc.Func != "DRAWICON" || dc.Type != 1 {
			t.Fatalf("[%s] DRAWICON+DRAWABOVE Command = %+v", name, dc)
		}
		if len(outs[0].SuffixParams) != 1 || outs[0].SuffixParams[0] != "DRAWABOVE" {
			t.Fatalf("[%s] DRAWICON SuffixParams = %v, want [DRAWABOVE]", name, outs[0].SuffixParams)
		}

		// DRAWNUMBER(C>10,LOW,C)：在最低价位置显示收盘价
		outs, err = ex.Run("DRAWNUMBER(C>10,LOW,C);", map[string][]float64{"C": {5, 15}, "LOW": {4, 14}})
		if err != nil {
			t.Fatalf("[%s] DRAWNUMBER: %v", name, err)
		}
		dc, ok = outs[0].Raw.(drawing.Command)
		if !ok {
			t.Fatalf("[%s] DRAWNUMBER Raw 类型 %T", name, outs[0].Raw)
		}
		if dc.Func != "DRAWNUMBER" {
			t.Fatalf("[%s] DRAWNUMBER Func = %q, want DRAWNUMBER", name, dc.Func)
		}
		if len(dc.Price) != 2 || dc.Price[0] != 4 || dc.Price[1] != 14 {
			t.Fatalf("[%s] DRAWNUMBER Price = %v, want [4 14]", name, dc.Price)
		}
		if len(dc.Data) != 2 || dc.Data[0] != 5 || dc.Data[1] != 15 {
			t.Fatalf("[%s] DRAWNUMBER Data = %v, want [5 15]", name, dc.Data)
		}
	}
}

// TestInBlock 验证 INBLOCK 数据源函数（依赖 SetBlockChecker + ctx.Symbol）。
func TestInBlock(t *testing.T) {
	ex := NewMaiExecutor()
	ex.SetBlockChecker(BlockCheckerFunc(func(symbol, block string) bool {
		return symbol == "000001.SZ" && block == "沪深300"
	}))
	outs, err := ex.RunWithContext("R:INBLOCK('沪深300');", Ctx{Symbol: "000001.SZ"}, nil)
	if err != nil {
		t.Fatalf("INBLOCK 命中: %v", err)
	}
	if got := outs[0].Raw; got != float64(1) {
		t.Fatalf("INBLOCK('沪深300') = %v, want 1", got)
	}

	outs, err = ex.RunWithContext("R:INBLOCK('中证500');", Ctx{Symbol: "000001.SZ"}, nil)
	if err != nil {
		t.Fatalf("INBLOCK 未命中: %v", err)
	}
	if got := outs[0].Raw; got != float64(0) {
		t.Fatalf("INBLOCK('中证500') = %v, want 0", got)
	}
}
