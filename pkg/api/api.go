package api

import (
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/lyr-2000/mylang/pkg/extensions/indicators"
	"github.com/lyr-2000/mylang/pkg/mylang"
	"github.com/spf13/cast"
)

// Ctx 是 mylang.Context 的宿主别名。
type Ctx = mylang.Context

// RankItem 是横截面排名中的一个个体，可以是股票，也可以是板块指数。
type RankItem struct {
	Key    string
	Values []float64
	Ts     []int64
}

// RankUniverse 是一次横截面统计的全体个体，Items 顺序决定“品种序号”。
type RankUniverse struct {
	Key   string
	Items []RankItem
}

// SectorResolver 由宿主实现：解析板块表达式，并给出该板块的指标输出矩阵。
type SectorResolver interface {
	ResolveUniverse(sectorExpr string, indicator string, outputIndex int, ctx Ctx) (RankUniverse, string, error)
}

type MaiExecutor struct {
	*mylang.MylangInterpreter
	PreCompiledProgram *mylang.Program //if not nil ,use it to execute the program
	DateTimeKey        string
	Resolver           SectorResolver
}

// OutputKind 输出变量的语义类型（uint8 减少内存占用）。
type OutputKind uint8

const (
	KindNaN    OutputKind = iota // 无效/空输出（无法识别类型或值无效），零值默认
	KindSeries                   // 一般数字序列（指标值，如 RSI/MACD）
	KindBool                     // 布尔序列（如 CROSS 的结果）
	KindBuy                      // 买入信号（BUY）
	KindSell                     // 卖出信号（SELL）
	KindDraw                     // 画图信号（预留：DRAWICON/DRAWLINE 等）
)

// Output 是一条画图变量的运行结果，顺序与脚本中 ":" 赋值顺序一致。
// Kind 标识语义类型；Values 统一承载结果（数字=值、布尔=0/1、信号=价位 NaN 表示无信号）；
// Raw 保留原始结果供宿主做扩展处理（布尔输出时仍保留 []bool）。
type Output struct {
	Name         string
	Order        int
	Kind         OutputKind
	SuffixParams []string
	Values       []float64
	Raw          any
}

// SetDateTimeKey 显式指定时间列变量名，避免 GetDateTimeArray 的猜测逻辑选错列。
func (m *MaiExecutor) SetDateTimeKey(name string) {
	m.DateTimeKey = name
}

// SetSectorResolver 注入板块数据提供者；INSORT/INSUM 依赖它。
func (m *MaiExecutor) SetSectorResolver(r SectorResolver) {
	m.Resolver = r
}

func (m *MaiExecutor) SetCustomVariableGetter(getter func(name string) any) {
	m.MylangInterpreter.Interp.CustomVariableGetter = getter
}

func (m *MaiExecutor) findOptionalDateTimeKey() string {
	var allKey = []string{
		"dateTime", "date", "Date", "DATE", "TS", "Ts", "ts", "timestamp", "time",
	}
	for _, key := range allKey {
		_, ok := m.MylangInterpreter.GetVariable(key)
		if ok {
			return key
		}
	}
	return ""
}

func (m *MaiExecutor) GetDateTimeArray() []any {
	if m.DateTimeKey == "" {
		dateKey := m.findOptionalDateTimeKey()
		if dateKey == "" {
			dateKey = "dateTime"
		}
		m.DateTimeKey = dateKey
	}
	d, ok := m.MylangInterpreter.GetVariable(m.DateTimeKey)
	if !ok {
		return nil
	}
	d1, _ := mylang.ToSlice(d)
	return d1
}

func NewMaiExecutor() *MaiExecutor {
	d := &MaiExecutor{
		MylangInterpreter: mylang.NewMylangInterpreter(),
	}
	d.registerFuncs()
	d.RegisterContextFunction("INSORT", d.insort)
	d.RegisterContextFunction("INSUM", d.insum)
	return d
}

func SetOutput(output io.Writer) {
	mylang.SetLogger(log.New(output, "", log.LstdFlags))
}

func Arrayfloat64(b any) []float64 {
	return cast.ToFloat64Slice(b)
}

func (b *MaiExecutor) GetBoolArray(name string) []bool {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	return cast.ToBoolSlice(d)
}

func (b *MaiExecutor) GetByteArray(name string) []byte {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	x, _ := cast.ToUint8SliceE(d)
	return x
}

func (b *MaiExecutor) GetFloat64Array(name string) []float64 {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	return Arrayfloat64(d)
}

func (b *MaiExecutor) GetVariableSlice(name string) []any {
	d, ok := b.MylangInterpreter.GetVariable(name)
	if !ok {
		return nil
	}
	d1, _ := mylang.ToSlice(d)
	return d1
}

func (b *MaiExecutor) SetVarMap(mp map[string][]float64) {
	for k, arr := range mp {
		b.SetVar(k, arr)
	}
}

// OPEN,HIGH,LOW,CLOSE,VOLUME,UnixSec -> O,H,L,C,V,Ts
func (b *MaiExecutor) SetVarNameAlias(alias map[string]string) {
	if alias == nil {
		if _, ok := b.MylangInterpreter.GetVariable("C"); ok {
			alias = map[string]string{
				"C": "CLOSE",
				"O": "OPEN",
				"L": "LOW",
				"H": "HIGH",
				"V": "VOLUME",
			}
		} else {
			alias = map[string]string{
				"OPEN":   "O",
				"HIGH":   "H",
				"LOW":    "L",
				"CLOSE":  "C",
				"VOLUME": "V",
			}
		}
	}
	for from, to := range alias {
		d, ok := b.MylangInterpreter.GetVariable(from)
		if ok {
			b.MylangInterpreter.SetVar(to, d)
		}
	}

}

func (m *MaiExecutor) CompileCode(code string) error {
	program, err := m.MylangInterpreter.Compile(code)
	if err != nil {
		m.PreCompiledProgram = nil
		return err
	}
	m.PreCompiledProgram = program
	return nil
}

// CompileSource 只编译并返回可缓存的语法树，不修改执行器的 PreCompiledProgram。
// 宿主可按脚本源码指纹缓存结果，配合 RunCompiled 复用。
func (m *MaiExecutor) CompileSource(source string) (*mylang.Program, error) {
	return m.MylangInterpreter.Compile(source)
}

func (m *MaiExecutor) PrintProgramTree() {
	if m.PreCompiledProgram == nil {
		fmt.Println("PreCompiledProgram is nil")
		return
	}

	fmt.Println("=== PreCompiledProgram AST ===")
	fmt.Printf("Total statements: %d\n", len(m.PreCompiledProgram.Statements))
	fmt.Println()

	for i, stmt := range m.PreCompiledProgram.Statements {
		fmt.Printf("Statement %d:\n", i)
		m.printStatement(stmt, 0)
		fmt.Println()
	}
	fmt.Println("=== End of AST ===")
}

func (m *MaiExecutor) printStatement(stmt mylang.Statement, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch s := stmt.(type) {
	case *mylang.AssignmentStatement:
		fmt.Printf("%sAssignmentStatement:\n", indentStr)
		fmt.Printf("%s  Name: %s\n", indentStr, s.Name.String())
		fmt.Printf("%s  IsDrawingVar: %t\n", indentStr, s.IsOutputVar)
		fmt.Printf("%s  Value:\n", indentStr)
		m.printExpression(s.Value, indent+2)

	case *mylang.ExpressionStatement:
		fmt.Printf("%sExpressionStatement:\n", indentStr)
		fmt.Printf("%s  Expression:\n", indentStr)
		m.printExpression(s.Expression, indent+2)

	default:
		fmt.Printf("%sUnknown statement type: %T\n", indentStr, stmt)
		fmt.Printf("%s  String: %s\n", indentStr, stmt.String())
	}
}

func (m *MaiExecutor) printExpression(expr mylang.Expression, indent int) {
	if expr == nil {
		fmt.Printf("%s<nil>\n", strings.Repeat("  ", indent))
		return
	}

	indentStr := strings.Repeat("  ", indent)

	switch e := expr.(type) {
	case *mylang.Identifier:
		fmt.Printf("%sIdentifier: %s\n", indentStr, e.Value)

	case *mylang.NumberLiteral:
		fmt.Printf("%sNumberLiteral: %f\n", indentStr, e.Value)

	case *mylang.StringLiteral:
		fmt.Printf("%sStringLiteral: %s\n", indentStr, e.Value)

	case *mylang.BinaryExpression:
		fmt.Printf("%sBinaryExpression:\n", indentStr)
		fmt.Printf("%s  Operator: %s\n", indentStr, e.Operator)
		fmt.Printf("%s  Left:\n", indentStr)
		m.printExpression(e.Left, indent+2)
		fmt.Printf("%s  Right:\n", indentStr)
		m.printExpression(e.Right, indent+2)

	case *mylang.FunctionCall:
		fmt.Printf("%sFunctionCall:\n", indentStr)
		fmt.Printf("%s  Function:\n", indentStr)
		m.printExpression(e.Function, indent+2)
		fmt.Printf("%s  Arguments (%d):\n", indentStr, len(e.Arguments))
		for i, arg := range e.Arguments {
			fmt.Printf("%s    [%d]:\n", indentStr, i)
			m.printExpression(arg, indent+3)
		}

	default:
		fmt.Printf("%sUnknown expression type: %T\n", indentStr, expr)
		fmt.Printf("%s  String: %s\n", indentStr, expr.String())
	}
}

func (m *MaiExecutor) ExecuteProgram() error {
	if m.PreCompiledProgram == nil {
		return fmt.Errorf("PreCompiledProgram is nil")
	}
	if len(m.PreCompiledProgram.Errors) > 0 {
		return &mylang.CompileError{Errors: m.PreCompiledProgram.Errors}
	}
	m.ResetRuntimeState()
	m.MylangInterpreter.ExecuteProgram(m.PreCompiledProgram)
	return m.Err
}

// RunCode 执行麦语言代码。已有预编译程序时返回错误，不再 panic。
func (m *MaiExecutor) RunCode(code string) (err error) {
	if m.PreCompiledProgram != nil {
		return fmt.Errorf("MaiExecutor 已有预编译程序，请使用 ExecuteProgram() 执行")
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("run code panic: %v", r)
		}
	}()
	m.ResetRuntimeState()
	m.Execute(code)
	return m.Err
}

// Run 一次性执行入口：重置执行状态、注入列式变量、编译运行并返回有序输出。
// 已注册的内置/自定义函数会由 Reset 保留，调用方无须关心解释器生命周期。
func (m *MaiExecutor) Run(source string, vars map[string][]float64) ([]Output, error) {
	return m.RunWithContext(source, Ctx{}, vars)
}

// RunWithContext 带执行上下文运行脚本。
// INSORT/INSUM 等横截面函数依赖 ctx.Symbol/Ts/DataVersion。
func (m *MaiExecutor) RunWithContext(source string, ctx Ctx, vars map[string][]float64) ([]Output, error) {
	m.PreCompiledProgram = nil
	m.Reset()
	m.SetContext(ctx)
	if len(vars) > 0 {
		m.SetVarMap(vars)
		m.SetVarNameAlias(nil)
	}
	if err := m.RunCode(source); err != nil {
		return nil, err
	}
	return m.Outputs(), nil
}

// RunCompiled 使用外部缓存的 Program 执行，适合高频请求场景。
func (m *MaiExecutor) RunCompiled(program *mylang.Program, vars map[string][]float64) ([]Output, error) {
	return m.RunCompiledWithContext(program, Ctx{}, vars)
}

// RunCompiledWithContext 带上下文复用预编译 Program，适合全市场批量执行。
func (m *MaiExecutor) RunCompiledWithContext(program *mylang.Program, ctx Ctx, vars map[string][]float64) ([]Output, error) {
	if program == nil {
		return nil, fmt.Errorf("program is nil")
	}
	if len(program.Errors) > 0 {
		return nil, &mylang.CompileError{Errors: program.Errors}
	}
	m.PreCompiledProgram = nil
	m.Reset()
	m.SetContext(ctx)
	if len(vars) > 0 {
		m.SetVarMap(vars)
		m.SetVarNameAlias(nil)
	}
	m.ResetRuntimeState()
	m.MylangInterpreter.ExecuteProgram(program)
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Outputs(), nil
}

// Outputs 按脚本声明顺序返回画图变量结果。
func (m *MaiExecutor) Outputs() []Output {
	vars := m.MylangInterpreter.OrderedOutputVariables()
	out := make([]Output, 0, len(vars))
	for _, v := range vars {
		raw, ok := m.MylangInterpreter.GetVariable(v.Name)
		if !ok {
			continue
		}
		kind, values := splitOutput(raw)
		out = append(out, Output{
			Name:         v.Name,
			Order:        v.Order,
			Kind:         kind,
			SuffixParams: v.SuffixParams,
			Values:       values,
			Raw:          raw,
		})
	}
	return out
}

// boolsToValues 把布尔序列转成 0/1 浮点序列。
func boolsToValues(bools []bool) []float64 {
	out := make([]float64, len(bools))
	for i, b := range bools {
		if b {
			out[i] = 1
		}
	}
	return out
}

func splitOutput(raw any) (OutputKind, []float64) {
	switch v := raw.(type) {
	case indicators.BuySignal:
		return KindBuy, []float64(v)
	case indicators.SellSignal:
		return KindSell, []float64(v)
	case []bool:
		return KindBool, boolsToValues(v)
	case []float64:
		return KindSeries, v
	case indicators.Series:
		return KindSeries, []float64(v)
	}
	items, ok := mylang.ToSlice(raw)
	if !ok {
		return KindNaN, nil
	}
	values := make([]float64, 0, len(items))
	bools := make([]bool, 0, len(items))
	for _, item := range items {
		if b, ok := item.(bool); ok {
			bools = append(bools, b)
			continue
		}
		values = append(values, cast.ToFloat64(item))
	}
	if len(items) > 0 && len(bools) == len(items) {
		return KindBool, boolsToValues(bools)
	}
	return KindSeries, values
}

func callBasicFunc(name string, args []any) (ret any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	b, err := indicators.CallIndicatorByReflection(name, args)
	return b, err
}

func (m *MaiExecutor) registerFuncs() {
	for _, name := range indicators.GetAllFuncNames() {
		funcName := name
		m.RegisterFunction(funcName, func(args []interface{}) interface{} {
			b, err := callBasicFunc(funcName, args)
			if err != nil {
				return fmt.Errorf("函数 %s: %w", funcName, err)
			}
			return b
		})
	}
}

// ---------------------------------------------------------------- 横截面统计

const defaultMaxBars = 6000

// RankTable 是一个板块在某指标某输出上的横截面结果，INSORT 与 INSUM 共享。
type RankTable struct {
	Ts         []int64
	Values     map[string][]float64
	Ranks      map[string][]float64
	Aggregates map[int][]float64
}

var rankCache sync.Map

func (m *MaiExecutor) getRankTable(u RankUniverse, indicator string, output int, order int, ctx Ctx) (*RankTable, error) {
	key := rankCacheKey(u.Key, indicator, output, order, ctx)
	if v, ok := rankCache.Load(key); ok {
		return v.(*RankTable), nil
	}
	table, err := buildRankTable(u, order, ctx)
	if err != nil {
		return nil, err
	}
	rankCache.Store(key, table)
	return table, nil
}

func rankCacheKey(universe, indicator string, output, order int, ctx Ctx) string {
	return universe + "|" + indicator + "|" + strconv.Itoa(output) + "|" +
		strconv.Itoa(order) + "|" + strconv.FormatInt(ctx.DataVersion, 10) + "|" +
		tsFingerprint(ctx.Ts)
}

func tsFingerprint(ts []int64) string {
	var h uint64 = 14695981039346656037
	for _, v := range ts {
		for i := 0; i < 8; i++ {
			h ^= uint64(byte(v >> (8 * i)))
			h *= 1099511628211
		}
	}
	return strconv.FormatUint(h, 16)
}

func buildRankTable(u RankUniverse, order int, ctx Ctx) (*RankTable, error) {
	table := &RankTable{
		Ts:         append([]int64(nil), ctx.Ts...),
		Values:     make(map[string][]float64, len(u.Items)),
		Ranks:      make(map[string][]float64, len(u.Items)),
		Aggregates: make(map[int][]float64, 6),
	}
	for calc := 0; calc <= 5; calc++ {
		table.Aggregates[calc] = nanSeries(len(ctx.Ts))
	}

	indexByTs := make(map[int64]int, len(ctx.Ts))
	for i, ts := range ctx.Ts {
		indexByTs[ts] = i
	}

	for _, item := range u.Items {
		vals := nanSeries(len(ctx.Ts))
		byTs := make(map[int64]float64, len(item.Ts))
		for i, ts := range item.Ts {
			if i < len(item.Values) {
				byTs[ts] = item.Values[i]
			}
		}
		for ts, v := range byTs {
			if idx, ok := indexByTs[ts]; ok {
				vals[idx] = v
			}
		}
		table.Values[item.Key] = vals
		table.Ranks[item.Key] = nanSeries(len(ctx.Ts))
	}

	valueAt := func(itemIdx, bar int) float64 {
		return table.Values[u.Items[itemIdx].Key][bar]
	}
	keyAt := func(itemIdx int) string { return u.Items[itemIdx].Key }

	for bar := range ctx.Ts {
		valid := make([]int, 0, len(u.Items))
		for j := range u.Items {
			if !math.IsNaN(valueAt(j, bar)) {
				valid = append(valid, j)
			}
		}
		if len(valid) == 0 {
			continue
		}

		rankOrder := append([]int(nil), valid...)
		sort.SliceStable(rankOrder, func(a, b int) bool {
			va, vb := valueAt(rankOrder[a], bar), valueAt(rankOrder[b], bar)
			if va != vb {
				if order == 0 {
					return va > vb
				}
				return va < vb
			}
			return keyAt(rankOrder[a]) < keyAt(rankOrder[b])
		})
		for rank, j := range rankOrder {
			table.Ranks[keyAt(j)][bar] = float64(rank + 1)
		}

		maxOrder := append([]int(nil), valid...)
		sort.SliceStable(maxOrder, func(a, b int) bool {
			va, vb := valueAt(maxOrder[a], bar), valueAt(maxOrder[b], bar)
			if va != vb {
				return va > vb
			}
			return keyAt(maxOrder[a]) < keyAt(maxOrder[b])
		})
		minOrder := append([]int(nil), valid...)
		sort.SliceStable(minOrder, func(a, b int) bool {
			va, vb := valueAt(minOrder[a], bar), valueAt(minOrder[b], bar)
			if va != vb {
				return va < vb
			}
			return keyAt(minOrder[a]) < keyAt(minOrder[b])
		})

		sum := 0.0
		for _, j := range valid {
			sum += valueAt(j, bar)
		}
		table.Aggregates[0][bar] = sum
		table.Aggregates[1][bar] = sum / float64(len(valid))
		table.Aggregates[2][bar] = valueAt(maxOrder[0], bar)
		table.Aggregates[3][bar] = valueAt(minOrder[0], bar)
		table.Aggregates[4][bar] = float64(maxOrder[0] + 1)
		table.Aggregates[5][bar] = float64(minOrder[0] + 1)
	}
	return table, nil
}

func nanSeries(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	return out
}

func toIntArg(v any) (int, bool) {
	i, err := cast.ToIntE(v)
	return i, err == nil
}

func (m *MaiExecutor) crossSectionError(ctx Ctx, err error) any {
	if ctx.Strict {
		return err
	}
	return nanSeries(len(ctx.Ts))
}

func (m *MaiExecutor) crossSectionGuard(ctx Ctx) any {
	maxBars := ctx.MaxBars
	if maxBars <= 0 {
		maxBars = defaultMaxBars
	}
	if len(ctx.Ts) == 0 || len(ctx.Ts) > maxBars {
		return nanSeries(len(ctx.Ts))
	}
	return nil
}

func (m *MaiExecutor) insort(ctx Ctx, args []any) any {
	if blocked := m.crossSectionGuard(ctx); blocked != nil {
		return blocked
	}
	if len(args) != 4 {
		return fmt.Errorf("INSORT 需要 4 个参数")
	}
	sector, ok := args[0].(string)
	if !ok || sector == "" {
		return fmt.Errorf("INSORT 板块名称必须是非空字符串")
	}
	indicator, ok := args[1].(string)
	if !ok || indicator == "" {
		return fmt.Errorf("INSORT 指标名称必须是非空字符串")
	}
	output, ok := toIntArg(args[2])
	if !ok || output <= 0 {
		return fmt.Errorf("INSORT 指标输出必须是正整数")
	}
	order, ok := toIntArg(args[3])
	if !ok || (order != 0 && order != 1) {
		return fmt.Errorf("INSORT 升降序只能是 0(降序) 或 1(升序)")
	}
	if m.Resolver == nil {
		return m.crossSectionError(ctx, fmt.Errorf("INSORT/INSUM 需要 SectorResolver"))
	}

	u, currentKey, err := m.Resolver.ResolveUniverse(sector, indicator, output, ctx)
	if err != nil {
		return m.crossSectionError(ctx, err)
	}
	table, err := m.getRankTable(u, indicator, output, order, ctx)
	if err != nil {
		return m.crossSectionError(ctx, err)
	}
	ranks, ok := table.Ranks[currentKey]
	if !ok {
		return m.crossSectionError(ctx, fmt.Errorf("INSORT 板块结果缺少当前品种 %s", currentKey))
	}
	return ranks
}

func (m *MaiExecutor) insum(ctx Ctx, args []any) any {
	if blocked := m.crossSectionGuard(ctx); blocked != nil {
		return blocked
	}
	if len(args) != 4 {
		return fmt.Errorf("INSUM 需要 4 个参数")
	}
	sector, ok := args[0].(string)
	if !ok || sector == "" {
		return fmt.Errorf("INSUM 板块名称必须是非空字符串")
	}
	indicator, ok := args[1].(string)
	if !ok || indicator == "" {
		return fmt.Errorf("INSUM 指标名称必须是非空字符串")
	}
	output, ok := toIntArg(args[2])
	if !ok || output <= 0 {
		return fmt.Errorf("INSUM 指标输出必须是正整数")
	}
	calcType, ok := toIntArg(args[3])
	if !ok || calcType < 0 || calcType > 5 {
		return fmt.Errorf("INSUM 计算类型只能是 0~5")
	}
	if m.Resolver == nil {
		return m.crossSectionError(ctx, fmt.Errorf("INSORT/INSUM 需要 SectorResolver"))
	}

	u, _, err := m.Resolver.ResolveUniverse(sector, indicator, output, ctx)
	if err != nil {
		return m.crossSectionError(ctx, err)
	}
	table, err := m.getRankTable(u, indicator, output, 0, ctx)
	if err != nil {
		return m.crossSectionError(ctx, err)
	}
	return table.Aggregates[calcType]
}
