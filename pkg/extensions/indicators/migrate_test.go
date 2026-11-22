package indicators

import (
	"math"
	"testing"
)

// assertSeries 断言两个 Series 相等（NaN 与 NaN 视为相等）。
func assertSeries(t *testing.T, name string, got, want Series) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s 长度 %d != %d", name, len(got), len(want))
	}
	for i := range want {
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got[i]) {
				t.Errorf("%s[%d] = %v, want NaN", name, i, got[i])
			}
			continue
		}
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %v, want %v", name, i, got[i], want[i])
		}
	}
}

func TestNumber(t *testing.T) {
	got := Number(Series{1.7, -2.3, math.NaN(), 0, 3.9})
	assertSeries(t, "Number", got, Series{1, -2, 0, 0, 3})
}

func TestREFX(t *testing.T) {
	s := Series{1, 2, 3, 4, 5}
	// 标量 N=2：result[i] = s[i+2]，越界为 NaN
	assertSeries(t, "REFX 标量", REFX(s, 2), Series{3, 4, 5, math.NaN(), math.NaN()})
	// 序列 N：逐位置偏移
	assertSeries(t, "REFX 序列", REFX(s, Series{0, 1, 2, 1, 0}), Series{1, 3, 5, 5, 5})
	// 负 N 取绝对值（未来方向）
	assertSeries(t, "REFX 负N取绝对值", REFX(s, -2), Series{3, 4, 5, math.NaN(), math.NaN()})
}

func TestBARSSINCE(t *testing.T) {
	s := []bool{false, true, false, true, false}
	// N=0 → 全历史：第一次 true 在 index=1
	assertSeries(t, "BARSSINCE 全历史", BARSSINCE(s, 0), Series{math.NaN(), 0, 1, 2, 3})
	// N=3 → 3 周期窗口（窗口不足时从序列起点缩小）
	assertSeries(t, "BARSSINCE N=3", BARSSINCE(s, 3), Series{math.NaN(), 0, 1, 2, 1})
}

func TestSUMBARS(t *testing.T) {
	x := Series{1, 2, 3, 4, 5}
	// 标量 A=6：从当前往前累加直到 >= 6
	assertSeries(t, "SUMBARS 标量", SUMBARS(x, 6), Series{math.NaN(), math.NaN(), 3, 2, 2})
	// 序列 A
	assertSeries(t, "SUMBARS 序列", SUMBARS(x, Series{1, 3, 5, 7, 9}), Series{1, 2, 2, 2, 2})
}

func TestDSMA(t *testing.T) {
	x := make(Series, 100)
	for i := range x {
		x[i] = 10 + math.Sin(float64(i)/5)*2
	}
	got := DSMA(x, 20)
	if len(got) != len(x) {
		t.Fatalf("DSMA 长度 %d != %d", len(got), len(x))
	}
	if got[0] != x[0] {
		t.Errorf("DSMA[0] = %v, want %v", got[0], x[0])
	}
	for i, v := range got {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("DSMA[%d] = %v 非有限值", i, v)
		}
	}
}

func TestBUYSELL(t *testing.T) {
	cond := Series{0, 1, 0, 1, 0}
	price := Series{10, 11, 12, 13, 14}

	buy := BUY(cond, price)
	assertSeries(t, "BUY", Series(buy), Series{math.NaN(), 11, math.NaN(), 13, math.NaN()})

	sell := SELL(cond, price)
	assertSeries(t, "SELL", Series(sell), Series{math.NaN(), 11, math.NaN(), 13, math.NaN()})
}
