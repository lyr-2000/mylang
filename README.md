# Mylang 解释器

一个用 Go 实现的“麦语言”解释器。它既能做单品种的时序指标计算，也支持通过宿主注入的板块数据实现通达信风格的横向统计（`INSORT` / `INSUM`）。

模块路径：`github.com/lyr-2000/mylang`

## 特性

- 变量管理、表达式求值、数组与标量混合运算。
- 内置 MyTT 指标函数：`MA/EMA/SMA/HHV/LLV/REF/CROSS/KDJ/MACD/...`，以及 `NZ`。
- 用户自定义 Go 函数，支持上下文感知函数。
- 有序输出：脚本中每个 `:` 赋值对应一条结果序列，顺序稳定。
- 编译产物可缓存复用，适合全市场批量计算。
- 横截面函数 `INSORT`（板块内排名）和 `INSUM`（板块内统计）。

## 包结构

| 包 | 用途 |
|---|---|
| `pkg/api` | 推荐入口。`MaiExecutor`、执行入口、输出结构、横截面统计、`SectorResolver` |
| `pkg/mylang` | 核心解释器：lexer、parser、环境、求值、上下文函数注册 |
| `pkg/extensions/indicators` | MyTT 指标函数与反射调用 |
| `pkg/extensions/tradingcharts` | 图表扩展，非核心依赖 |

---

## 语言速览

### 赋值与输出

```
中间变量:=CLOSE+1;    // 不输出到结果里
MA5:MA(CLOSE,5);      // 输出到结果里，变量名即序列名
K:MA(CLOSE,5),COLORRED,NODRAW;  // 可带修饰符
```

- `:=` 是普通赋值。
- `:` 是画图/输出赋值，会进入 `Outputs()` 的结果。
- `:` 后面的修饰符会进入 `Output.SuffixParams`，由宿主决定如何渲染。

### 注释

支持两种注释：

```
{ 块注释，可嵌套 }
// 行注释，脚本首行可以写 //@title 等宿主元数据
```

单引号字符串里的内容会被原文保留，例如 `NAMELIKE('S//T')`。

### 常用内置函数

`ABS, REF, MA, EMA, SMA, WMA, HHV, LLV, STD, SUM, CROSS, COUNT, IF, NZ, MACD, KDJ, RSI, BOLL, ...`

完整列表见 `pkg/extensions/indicators/MyTT.go` 的 `functionMap`。

### KDJ 示例

```
RSV:=(CLOSE-LLV(LOW,9))/(HHV(HIGH,9)-LLV(LOW,9))*100;
K:SMA(RSV,3,1);
D:SMA(K,3,1);
J:3*K-2*D;
```

注意 `SMA` 是三参数 `SMA(S,N,M)`，与 Pine 的两参数 `sma(x,n)` 不同。

---

## 单品种时序 API

### 最简单用法

```go
package main

import (
	"fmt"
	"log"

	"github.com/lyr-2000/mylang/pkg/api"
)

func main() {
	ex := api.NewMaiExecutor()

	source := `
MA5:MA(CLOSE,5);
UP:CLOSE>MA(CLOSE,5);
`

	outs, err := ex.Run(source, map[string][]float64{
		"OPEN":   []float64{1, 2, 3, 4, 5, 6},
		"HIGH":   []float64{2, 3, 4, 5, 6, 7},
		"LOW":    []float64{0, 1, 2, 3, 4, 5},
		"CLOSE":  []float64{1, 2, 3, 4, 5, 6},
		"VOLUME": []float64{10, 10, 10, 10, 10, 10},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, o := range outs {
		fmt.Println(o.Name, o.Order, o.Kind, o.Values, o.SuffixParams)
	}
}
```

`NewMaiExecutor()` 会自动注册所有内置指标函数，以及 `INSORT` / `INSUM`。

### Output 结构

```go
type Output struct {
	Name         string    // ":" 赋值的变量名
	Order        int       // 声明顺序，从 1 开始
	Kind         OutputKind // KindNaN=无效 / KindSeries=数字 / KindBool=布尔 / KindBuy=买入 / KindSell=卖出 / KindDraw=画图
	SuffixParams []string  // COLORRED / NODRAW 等
	Values       []float64 // 数字=值；布尔=0/1；信号=价位（NaN 表示无信号）
	Raw          any       // 原始结果（布尔输出时仍保留 []bool）
}
```

`Values` 的下标与传入的 K 线数组下标一致。宿主需要自己保存时间戳数组，用它来生成图表点：

```go
for _, o := range outs {
	for i, v := range o.Values {
		if math.IsNaN(v) {
			continue
		}
		point := indicator.Point{Timestamp: bars.Ts[i], Value: v}
		_ = point
	}
}
```

### 输入变量约定

推荐用 `SetVarMap` 的形状：

```go
vars := map[string][]float64{
	"OPEN": bars.Open, "HIGH": bars.High,
	"LOW": bars.Low, "CLOSE": bars.Close,
	"VOLUME": bars.Vol,
}
```

`Run` 内部会调用 `SetVarNameAlias(nil)`，自动生成短别名：

```
CLOSE -> C
OPEN  -> O
HIGH  -> H
LOW   -> L
VOLUME -> V
```

时间列可以显式指定：

```go
ex.SetDateTimeKey("dateTime")
ex.SetVar("dateTime", bars.Ts)
```

### 复用编译产物

高频场景先编译，再多次执行：

```go
program, err := ex.CompileSource(source)
if err != nil {
	log.Fatal(err)
}

for _, vars := range allSymbols {
	outs, err := ex.RunCompiled(program, vars)
	if err != nil {
		log.Fatal(err)
	}
	_ = outs
}
```

`CompileSource` 不修改执行器状态，`Program` 可以按“源码指纹”缓存。

### 注册自定义函数

普通函数：

```go
ex.RegisterFunction("FIRST", func(args []interface{}) interface{} {
	arr := api.Arrayfloat64(args[0])
	if len(arr) == 0 {
		return 0.0
	}
	return arr[0]
})
```

上下文函数（可以读当前品种/时间轴）：

```go
ex.RegisterContextFunction("MYCTX", func(ctx api.Ctx, args []interface{}) interface{} {
	_ = ctx.Symbol
	return float64(len(ctx.Ts))
})
```

### 自定义变量提供者

```go
ex.SetCustomVariableGetter(func(name string) any {
	if name == "NOWDAY" {
		return time.Now().Format("20060102")
	}
	return nil
})
```

### 日志

默认日志输出到 `io.Discard`。全局修改：

```go
api.SetOutput(os.Stderr)
```

注意这是全局设置，会影响所有执行器。

---

## 横截面统计：INSORT / INSUM

这一部分用于“全市场/板块内选股”。语法仍然是单品种脚本，但函数内部会按板块展开成矩阵。

### 函数签名

```text
INSORT(板块名称, 指标名称, 指标输出, 升降序)
INSUM(板块名称, 指标名称, 指标输出, 计算类型)
```

例子：

```
排名:INSORT('软件服务','KDJ',3,0);
累加:INSUM('软件服务','KDJ',3,0);
均值:INSUM('软件服务','KDJ',3,1);
最大:INSUM('软件服务','KDJ',3,2);
最小:INSUM('软件服务','KDJ',3,3);
最大序号:INSUM('软件服务','KDJ',3,4);
最小序号:INSUM('软件服务','KDJ',3,5);
```

`3` 表示 KDJ 的第 3 个 `:` 输出，即 J。

### INSORT 语义

| 行为 | 说明 |
|---|---|
| 返回值 | 当前品种在板块中的排名时序，长度与 K 线一致 |
| 升降序 | `0=降序`，`1=升序`；降序时最大值排第 1 |
| 排名基准 | 从 1 开始，连续不跳号 |
| 并列处理 | 先按指标值排序，并列时按股票代码字符串字典序升序打破 |
| 输出序号 | 必须是正整数 |
| 当前 bar 无值 | 返回 NaN |
| bar 数超过 6000 | 整条序列返回 NaN，不计算 |

### INSUM 计算类型

| 类型 | 含义 |
|---|---|
| 0 | 累加 |
| 1 | 平均数 |
| 2 | 最大值 |
| 3 | 最小值 |
| 4 | 最大值所处的品种序号 |
| 5 | 最小值所处的品种序号 |

“品种序号”从 1 开始，对应 `RankUniverse.Items` 中的顺序；并列时按股票代码字典序决定取哪一个。

INSUM 返回的是一条对所有成员都相同的序列。

### 上下文 Ctx

```go
type Ctx struct {
	Symbol      string // 当前品种代码
	Ts          []int64 // 时间轴，用于横截面对齐
	DataVersion int64  // 行情数据版本，缓存 key 的一部分
	Strict      bool   // true: 解析失败返回错误；false: 返回 NaN
	MaxBars     int    // 默认 6000
}
```

带上下文执行：

```go
ctx := api.Ctx{
	Symbol:      "600519",
	Ts:          bars.Ts,
	DataVersion: version,
	Strict:      true,
}

outs, err := ex.RunWithContext(source, ctx, vars)
```

预编译 + 上下文：

```go
program, _ := ex.CompileSource(source)
outs, err := ex.RunCompiledWithContext(program, ctx, vars)
```

### 宿主实现 SectorResolver

mylang 不内置板块成员和指标注册表，宿主需要实现：

```go
type RankItem struct {
	Key    string    // 股票代码或板块指数标识
	Values []float64 // 该 item 的指标输出时序
	Ts     []int64
}

type RankUniverse struct {
	Key   string      // 缓存键，必须稳定
	Items []RankItem  // 顺序决定“品种序号”
}

type SectorResolver interface {
	ResolveUniverse(
		sectorExpr string,
		indicator string,
		outputIndex int,
		ctx Ctx,
	) (RankUniverse, string, error)
}
```

返回值中的 `string` 是“当前品种在 universe 中的 key”。普通板块场景下通常是 `ctx.Symbol`；`HYINSORT` 这类板块指数排名场景下，应是当前股所属板块指数的 key。

注入：

```go
ex.SetSectorResolver(myResolver)
```

未设置 Resolver 时调用 `INSORT/INSUM`：`Strict=true` 返回错误，否则返回 NaN。

### SectorResolver 的职责

mylang 会把 `sectorExpr` 原样传给 resolver。resolver 负责解析：

- 系统板块：上证Ａ股、深证Ａ股、沪深Ａ股/全部Ａ股、创业板、科创板、北证Ａ股、自选股、板块指数等。
- 前缀：`HY.` 行业、`GN.` 概念、`MY.` 自定义板块。
- `HYBLOCK`：当前股所属行业板块。
- `HYINSORT`：所有行业板块指数；当前项是当前股所属行业指数。
- `GNINSORT.n` / `FGINSORT.n`：所有概念/风格板块指数。
- `MYGNBK` / `MYFGBK`：当前股所属的概念/风格板块指数集合，常用于 INSUM。
- `CLOUDGPS`：云端板块，resolver 可返回 unsupported。

指标名与输出序号也由 resolver 映射到具体脚本和变量：

```text
('KDJ', 3) -> KDJ 脚本的第 3 个 ":" 输出，一般是 J
```

### 对齐与缓存

- 成员指标按 `Ts` 对齐，不按下标对齐，停牌/新股天然缺失。
- 排名表按 `(universeKey, indicator, output, order, dataVersion, tsFingerprint)` 缓存。
- `INSORT` 与 `INSUM` 共享同一张 `RankTable`。
- 当前实现是进程级 `sync.Map` 缓存，没有自动淘汰；数据版本变化会生成新 key。

---

## 生命周期与并发

- 建议一次执行新建一个 `MaiExecutor`，或按 goroutine 池化。
- `Reset()` 会清理变量和输出状态，但保留注册函数、`CustomVariableGetter` 和 `Context`。
- `ResetRuntimeState()` 只清理一次执行产生的输出映射和错误，适合复用 `Program`。
- 解释器本身有状态；不要多个 goroutine 共用同一个执行器。
- 全局 `Logger` 默认丢弃输出；批量并发时不要调用 `SetOutput` 改成共享 writer。

---

## 从旧 API 迁移

旧写法仍然可用：

```go
ex := api.NewMaiExecutor()
ex.SetVar("CLOSE", closes)
err := ex.RunCode("MA5:MA(CLOSE,5);")
```

但注意：

- `CompileCode` + `ExecuteProgram` 中的 panic 已改为返回错误。
- `RunCode` 在已有 `PreCompiledProgram` 时返回错误，不再 panic。
- 需要有序结果时用 `ex.Outputs()`，不要依赖 map 遍历顺序。
- 需要横截面能力时用 `RunWithContext` / `RunCompiledWithContext`。
