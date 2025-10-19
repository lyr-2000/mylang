package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

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
	executor.CompileCode(`
MA22:MA(CLOSE,22);
MA55:MA(CLOSE,55);
MA144:MA(CLOSE,144);
RSI$G2:RSI(CLOSE,14);
VOL22:MA(VOLUME,22)

	`)
	executor.ExecuteProgram()
	if executor.Err != nil {
		log.Panic(executor.Err)
		return
	}

	chart := charts.NewKlineChart(executor, "test")
	chart.SetTickFormatType(charts.TickFormatTypeDateTime)
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

	for varName := range executor.GetDrawingVariables() {
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
