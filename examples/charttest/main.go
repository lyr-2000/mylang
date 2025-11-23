//go:build ignore
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/charts"
	grob "github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/generated/v2.34.0/graph_objects"
	"github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/pkg/types"
)

type Klinedef struct {
	// {"version":0,"originKl":null,"klines":[{"o
	Klines []*charts.Ohlcv
}

func S(s string) types.StringType {
	return types.S(s)
}

// 加载k线，执行指标，并且生成图表
func main() {
	executor := api.NewMaiExecutor()
	b, err := os.ReadFile("examples/charttest/000001.SZ.json")
	if err != nil {
		panic(err)
	}
	var bdef Klinedef

	json.Unmarshal(b, &bdef)
	var mp = make(map[string][]float64)
	var datetime []any
	sort.Slice(bdef.Klines, func(i, j int) bool {
		return bdef.Klines[i].GetAny("date").(string) < bdef.Klines[j].GetAny("date").(string)
	})
	for _, kl := range bdef.Klines {
		mp["CLOSE"] = append(mp["CLOSE"], kl.GetFloat64("close"))
		mp["OPEN"] = append(mp["OPEN"], kl.GetFloat64("open"))
		mp["HIGH"] = append(mp["HIGH"], kl.GetFloat64("high"))
		mp["LOW"] = append(mp["LOW"], kl.GetFloat64("low"))
		mp["VOLUME"] = append(mp["VOLUME"], kl.GetFloat64("vol"))
		datetime = append(datetime, kl.Store["date"])
	}
	for name, data := range mp {
		executor.SetVar(name, data)
	}
	executor.SetVar("dateTime", datetime)

	executor.SetVarNameAlias(map[string]string{
		"CLOSE":    "C",
		"OPEN":     "O",
		"HIGH":     "H",
		"LOW":      "L",
		"VOLUME":   "V",
		"dateTime": "Ts",
	})
	executor.SetVar("$abcd_1好",16)
	// executor.RunCode("abc你好456:=$abcd_1好")
	err = executor.CompileCode(`
MA22:MA(CLOSE,22);
MA55:MA(CLOSE,55);
MA144:MA(CLOSE,144);
RSI$G2:RSI(CLOSE,14);
VOL22:MA(VOLUME,22);
ZTPPrice:ZTPRICE(C,0.1);
涨停:=C>=ZTPRICE(REF(C,1),0.1);

去除:=1;
ZTPPrice:ZTPRICE(C,0.1);
涨停:=C>=ZTPRICE(REF(C,1),0.1);
前天涨停:=REF(涨停,2);
昨天阴线:=REF(C,1)<REF(O,1);
碰涨停:=H>=ZTPRICE(REF(C,1),0.1);
N字涨停板$m:=前天涨停 AND 昨天阴线 AND 碰涨停 AND 去除;
a:codelike(x);
	`)
	if err != nil {
		log.Panic(err)
		return
	}
	executor.PrintProgramTree()
	err = executor.ExecuteProgram()
	if err != nil {
		log.Panic(err)
		return
	}
	if executor.Err != nil {
		log.Panic(executor.Err)
		return
	}

	chart := charts.NewKlineChart(executor, "test")
	chart.SetTickFormatType(charts.TickFormatTypeDateTime)
	chart.KlineColorMode = charts.GreenDownAndRedUp
	chart.SetDefaultKlineChart()
	var x []types.StringType
	for i := 0; i < 30; i++ {
		x = append(x, types.StringType(fmt.Sprintf("buy %d", i)))
	}
	log.Println(executor.GetVariable("abc你好456"))
	chart.AddMarker(0, "buy", types.DataArray(datetime[:30]),
		types.DataArray(mp["CLOSE"][:30]),
		x, &grob.ScatterMarker{
			Color: types.ArrayOKValue(types.UseColor("red")),
			Size:  types.ArrayOKValue(types.N(10)),
		})
	
	x1,_ := executor.GetOutputVariable(1)
	x2,_ := executor.GetVariable("MA22")
	if reflect.DeepEqual(x1,x2) {
		fmt.Println("OK")
		// log.Fatal("xxxxxxx variable error ")
		fmt.Println(x1)
		fmt.Println(x2)
	}
	b1,_ := executor.GetLastOutput()

	_ = b1
	
	xxx := executor.GetFloat64Array("N字涨停板$m")
	_ = xxx
	xxx = executor.GetFloat64Array("碰涨停")
	_ = xxx
	xxx = executor.GetFloat64Array("昨天阴线")
	_ = xxx
	for varName := range executor.GetOutputVariableMap() {
		if varName == "RSI$G2" || varName == "VOL22" {
			continue
		}
		chart.AddMainCharts(&grob.Scatter{
			// Uid:       types.StringType(varName),
			Name:      charts.S(varName),
			X:         charts.Array(executor.GetDateTimeArray()),
			Y:         charts.Array(executor.GetFloat64Array(varName)),
		})
	}
	chart.AddCharts(1,&grob.Scatter{
		Name:      charts.S("VOL22"),
		X:         charts.Array(executor.GetDateTimeArray()),
		Y:         charts.Array(executor.GetFloat64Array("VOL22")),
		Hovertext: charts.HoverTextArray(charts.StringRepeatToArray("VOL22 Hover", len(executor.GetDateTimeArray()))...),
		Xaxis:     charts.S(charts.Xaxis()),
		Yaxis:     charts.S(charts.Yaxis(1)),
	})
	chart.AddCharts(2,&grob.Scatter{
		Name:      charts.S("RSI_70"),
		X:         charts.Array(executor.GetDateTimeArray()),
		Y:         charts.Array(charts.FloatRepeatToArray(70, len(executor.GetDateTimeArray()))),
		Xaxis:     charts.S(charts.Xaxis()),
		Yaxis:     charts.S(charts.Yaxis(2)),
	})
	chart.AddCharts(2,&grob.Scatter{
		Name:      charts.S("RSI_30"),
		X:         charts.Array(executor.GetDateTimeArray()),
		Y:         charts.Array(charts.FloatRepeatToArray(30, len(executor.GetDateTimeArray()))),
		Xaxis:     charts.S(charts.Xaxis()),
		Yaxis:     charts.S(charts.Yaxis(2)),
	})
	chart.AddCharts(2,&grob.Scatter{
		Name:      charts.S("RSI"),
		X:         charts.Array(executor.GetDateTimeArray()),
		Y:         charts.Array(executor.GetFloat64Array("RSI$G2")),
		Hovertext: charts.HoverTextArray(charts.StringRepeatToArray("RSI HoverInfo", len(executor.GetDateTimeArray()))...),
		Xaxis:     charts.S(charts.Xaxis()),
		Yaxis:     charts.S(charts.Yaxis(2)),
	})
	chart.AsHtml("test.html")
}
