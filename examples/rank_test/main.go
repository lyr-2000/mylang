package main

import (
	"fmt"
	"math"

	"github.com/lyr-2000/mylang/pkg/api"
)

type fakeResolver struct{}

func (f *fakeResolver) ResolveUniverse(sector, indicator string, output int, ctx api.Ctx) (api.RankUniverse, string, error) {
	if sector != "软件服务" || indicator != "KDJ" || output != 3 {
		return api.RankUniverse{}, "", fmt.Errorf("unexpected args: %q %q %d", sector, indicator, output)
	}

	patterns := map[string][]float64{
		"000001": {10, 40, 30, 20},
		"000002": {30, 20, 10, 40},
		"600519": {20, 30, 40, 10},
		"600036": {40, 10, 20, 30},
	}
	codes := []string{"000001", "000002", "600519", "600036"}

	items := make([]api.RankItem, 0, len(codes))
	for _, code := range codes {
		seq := patterns[code]
		values := make([]float64, len(ctx.Ts))
		for i := range ctx.Ts {
			values[i] = seq[i%len(seq)]
		}
		items = append(items, api.RankItem{
			Key:    code,
			Values: values,
			Ts:     ctx.Ts,
		})
	}
	return api.RankUniverse{
		Key:   "软件服务|KDJ|3",
		Items: items,
	}, ctx.Symbol, nil
}

func assertSeries(name string, got, want []float64) {
	if len(got) != len(want) {
		panic(fmt.Sprintf("%s len = %d, want %d: %v", name, len(got), len(want), got))
	}
	for i := range want {
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got[i]) {
				panic(fmt.Sprintf("%s[%d] = %v, want NaN", name, i, got[i]))
			}
			continue
		}
		if got[i] != want[i] {
			panic(fmt.Sprintf("%s[%d] = %v, want %v", name, i, got[i], want[i]))
		}
	}
}

func main() {
	ex := api.NewMaiExecutor()
	ex.SetSectorResolver(&fakeResolver{})

	ts := []int64{1, 2, 3, 4}
	vars := map[string][]float64{
		"OPEN":   {1, 2, 3, 4},
		"HIGH":   {2, 3, 4, 5},
		"LOW":    {0, 1, 2, 3},
		"CLOSE":  {1, 2, 3, 4},
		"VOLUME": {10, 10, 10, 10},
	}
	ctx := api.Ctx{
		Symbol:      "600519",
		Ts:          ts,
		DataVersion: 1,
		Strict:      true,
	}

	source := `
R:INSORT('软件服务','KDJ',3,0);
S:INSUM('软件服务','KDJ',3,0);
MXI:INSUM('软件服务','KDJ',3,4);
`
	outs, err := ex.RunWithContext(source, ctx, vars)
	if err != nil {
		panic(err)
	}
	if len(outs) != 3 {
		panic(fmt.Sprintf("outputs = %d, want 3", len(outs)))
	}

	// 600519 的 J 是 [20,30,40,10]，在全板块降序排名为 [3,2,1,4]。
	assertSeries("INSORT", outs[0].Values, []float64{3, 2, 1, 4})
	// 每根 bar 上四只股票的 J 之和都是 100。
	assertSeries("INSUM-sum", outs[1].Values, []float64{100, 100, 100, 100})
	// 最大值所在的品种序号：600036 / 000001 / 600519 / 000002。
	assertSeries("INSUM-max-index", outs[2].Values, []float64{4, 1, 3, 2})

	fmt.Println("INSORT/INSUM self-check ok")
}
