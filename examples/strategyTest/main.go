//go:build ignore
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/lyr-2000/mylang/pkg/api"
)

type Klinedef struct {
	// {"version":0,"originKl":null,"klines":[{"o
	Klines []*Ohlcv
}

// Ohlcv 是 K 线的宽松 JSON 解析辅助（原 charts.Ohlcv 内联，避免依赖图表库）。
type Ohlcv struct {
	Store map[string]any
}

func (b *Ohlcv) GetAny(name string) any {
	return b.Store[name]
}

func (b *Ohlcv) GetFloat64(name string) float64 {
	v, _ := b.Store[name].(float64)
	return v
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
