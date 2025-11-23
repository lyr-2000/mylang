//go:build ignore
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/lyr-2000/mylang/pkg/api"
	"github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/charts"
	// grob "github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/generated/v2.34.0/graph_objects"
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
	// executor.RunCode("abc你好456:=$abcd_1好")

	fmt.Println("TODO")
}
