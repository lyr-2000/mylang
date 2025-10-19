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
	executor.RunCode("abc你好456:=$abcd_1好")

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
	chart.AsHtml("test.html")
}
