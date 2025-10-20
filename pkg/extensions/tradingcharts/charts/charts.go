package charts

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/lyr-2000/mylang/pkg/api"
	grob "github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/generated/v2.34.0/graph_objects"
	"github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/pkg/offline"
	"github.com/lyr-2000/mylang/pkg/extensions/tradingcharts/go-plotly/pkg/types"
	"github.com/spf13/cast"
)

type ChartLib struct {
}

type Ohlcv struct {
	Store map[string]any
}

func (b *Ohlcv) GetFloat64(name string) float64 {
	return cast.ToFloat64(b.Store[name])
}

func (r *Ohlcv) UnmarshalJSON(data []byte) error {
	var mp map[string]any
	err := json.Unmarshal(data, &mp)
	if err != nil {
		return err
	}
	r.Store = mp
	return nil
}

type ChartSetting struct {
	Kline          []Ohlcv     `json:"kline"`
	MainIndicators []Indicator `json:"main_indicators"`
	SubCharts      []SubChart  `json:"sub_charts"`
}
type Indicator struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Values [][]float64 `json:"values"`
}
type SubChart struct {
	Name       string      `json:"name"`
	Indicators []Indicator `json:"indicators"`
}
type Settings struct {
	ChartSettings map[string]string
}

func (r *Settings) GetOrDefault(key string, defaultValue string) string {
	if r == nil || r.ChartSettings[key] == "" {
		return defaultValue
	}
	return r.ChartSettings[key]
}

type KlineChart struct {
	Title          string
	Tmp            map[int][]types.Trace
	Executor       *api.MaiExecutor
	Settings       *Settings
	TickFormatType string
	// Fig      *grob.Fig
	// KlineAlias map[string]string
}

var (
	TickFormatTypeDateNum  = "%Y%m%d"
	TickFormatTypeDate     = "%Y-%m-%d"
	TickFormatTypeDateTime = "%Y-%m-%d %H:%M"
)

func (r *KlineChart) SetTickFormatType(tickFormatType string) {
	if tickFormatType == "" {
		r.TickFormatType = TickFormatTypeDateTime
	} else {
		r.TickFormatType = tickFormatType
	}
}

type FigSettingOpt func(b *grob.Fig)

func (r *KlineChart) LayoutObj(title string) *grob.Fig {
	if r.TickFormatType == "" {
		r.SetTickFormatType(TickFormatTypeDateTime)
	}
	return &grob.Fig{
		Layout: &grob.Layout{
			Title: &grob.LayoutTitle{
				Text: types.StringType(title),
			},
			Xaxis: &grob.LayoutXaxis{
				Rangeslider: &grob.LayoutXaxisRangeslider{
					Visible: types.False,
				},
				Tickmode:   "auto",
				Tickformat: types.StringType(r.Settings.GetOrDefault("xaxis_tickformat", r.TickFormatType)),
			},
			XAxis2: &grob.LayoutXaxis{
				Rangeslider: &grob.LayoutXaxisRangeslider{
					Visible: types.False,
				},
				Tickmode:   "auto",
				Tickformat: types.StringType(r.Settings.GetOrDefault("xaxis2_tickformat", r.TickFormatType)),
			},
			// Yaxis: &grob.LayoutYaxis{
			// 	// Tickformat:  ".6",
			// 	// Hoverformat: ",.2r",
			// 	//tickFormat settings  https://community.plotly.com/t/how-to-change-the-way-large-numbers-are-displayed/72363
			// },
			Autosize: types.True,
			Height:   types.N(904),
			Calendar: grob.LayoutCalendarChinese,
			Grid: &grob.LayoutGrid{
				Rows:    types.I(2),
				Columns: types.I(1),
				Pattern: "independent",
			},
		},
	}
}

func NewKlineChart(executor *api.MaiExecutor, title string) *KlineChart {
	return &KlineChart{
		Executor: executor,
		// Fig:      layoutX(title),
	}
}

func (r *KlineChart) AsHtml(writePath string, opts ...FigSettingOpt) {
	w := r.ObjInit(opts...)
	offline.ToHtml(w, writePath)
}

func (r *KlineChart) SetDefaultKlineChart() {
	closeData := r.Executor.GetFloat64Array("C")
	openData := r.Executor.GetFloat64Array("O")
	highData := r.Executor.GetFloat64Array("H")
	lowData := r.Executor.GetFloat64Array("L")
	volumeData := r.Executor.GetFloat64Array("V")
	date := r.Executor.GetVariableSlice("dateTime")

	r.AddCharts(0,
		&grob.Candlestick{
			Uid:       "1",
			Name:      types.S("Kline"),
			Close:     types.DataArray(closeData),
			Open:      types.DataArray(openData),
			High:      types.DataArray(highData),
			Low:       types.DataArray(lowData),
			X:         types.DataArray(date),
			Xaxis:     types.S("x"),
			Yaxis:     types.S("y"),
			Hovertext: types.ArrayOKArray[types.StringType](),
		})
	r.AddCharts(1,
		&grob.Bar{
			Uid:       "2",
			Name:      types.S("Volume"),
			X:         types.DataArray(date),
			Y:         types.DataArray(volumeData),
			Xaxis:     types.S("x"),
			Yaxis:     types.S("y2"),
			Hovertext: types.ArrayOKArray[types.StringType](),
		})

}
func S(s string) types.StringType {
	return types.S(s)
}

func StringRepeatToArray(text string, count int) []string {
	var arr []string
	for i := 0; i < count; i++ {
		arr = append(arr, text)
	}
	return arr
}

func FloatRepeatToArray(text float64, count int) []float64 {
	var arr []float64
	for i := 0; i < count; i++ {
		arr = append(arr, text)
	}
	return arr
}

func HoverTextArray(texts ...string) *types.ArrayOK[*types.StringType] {
	var arr []*types.StringType
	for _, text := range texts {
		d := types.StringType(text)
		arr = append(arr, &d)
	}
	return &types.ArrayOK[*types.StringType]{Array: arr}
}

func ArrayOmitZero(b any) *types.DataArrayType {
	return types.DataArray[any](array(b, true))
}

func Array(b any) *types.DataArrayType {
	return types.DataArray[any](array(b, false))
}

func array(b any, omitZero bool) []any {
	switch b := b.(type) {
	case []float64:
		return copySlice(b, omitZero)
	case []string:
		return copySlice(b, omitZero)
	case []any:
		return copySlice(b, omitZero)
	}
	return nil
}

func copySlice[T any](x []T, omit0 bool) []any {
	var mp = make([]any, len(x))
	for i, v := range x {
		mp[i] = v

		f, ok := mp[i].(float64)
		if ok {
			if math.IsNaN(f) {
				mp[i] = nil
			}
			if omit0 && f == 0 {
				mp[i] = nil
			}
		}
	}
	return mp
}

func (rb *KlineChart) AddMarker(idx int, name string, x *types.DataArrayType, y *types.DataArrayType,
	hoverText []types.StringType, marker *grob.ScatterMarker) {
	r := &grob.Scatter{
		X:         x,
		Y:         y,
		Xaxis:     types.S("x"),
		Yaxis:     types.S("y"),
		Name:      types.StringType(name),
		Marker:    marker,
		Mode:      grob.ScatterModeMarkers,
		Hovertext: types.ArrayOKArray[types.StringType](hoverText...),
	}
	rb.AddCharts(idx, r)
}

func (r *KlineChart) ObjInit(opts ...FigSettingOpt) *grob.Fig {
	fig := r.LayoutObj(r.Title)
	type tmp struct {
		w      int
		slices []types.Trace
	}
	var tmpx []tmp
	for w, chart := range r.Tmp {
		tmpx = append(tmpx, tmp{w: w, slices: chart})
	}
	sort.Slice(tmpx, func(i, j int) bool {
		return tmpx[i].w < tmpx[j].w
	})
	var subplots [][]string
	for _, slot := range tmpx {
		w := slot.w
		fig.AddTraces(slot.slices...)
		x := "xy"
		if w != 0 {
			x = x + cast.ToString(w+1)
		}
		// r.Fig.Layout.Grid.Subplots = append(r.Fig.Layout.Grid.Subplots, []string{x})
		subplots = append(subplots, []string{x})
	}
	fig.Layout.Grid.Rows = types.I(len(tmpx))
	fig.Layout.Grid.Columns = types.I(1)
	fig.Layout.Grid.Subplots = subplots
	for _, opt := range opts {
		opt(fig)
	}
	return fig
}

func (r *KlineChart) AddMainCharts(charts ...types.Trace) {
	r.AddCharts(0, charts...)
}

func Xaxis() string {
	return "x"
}
func Yaxis(idx int) string {
	return "y" + cast.ToString(idx+1)
}

func (r *KlineChart) AddCharts(idx int, charts ...types.Trace) {
	if r.Tmp == nil {
		r.Tmp = make(map[int][]types.Trace)
	}
	r.Tmp[idx] = append(r.Tmp[idx], charts...)
}
