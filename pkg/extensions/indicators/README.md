# MyTT Go版本 - 技术指标函数库

这是MyTT Python版本技术指标函数库的Go语言实现，提供了丰富的技术分析指标函数。

## 功能特性

- **0级核心工具函数**: 数学运算、序列处理
- **1级应用层函数**: 条件判断、统计分析
- **2级技术指标函数**: MACD、KDJ、RSI、BOLL等经典指标
- **高级函数版本**: SAR、薛斯通道等复杂指标

## 主要指标

### 移动平均类
- `MA(S, N)` - 简单移动平均
- `EMA(S, N)` - 指数移动平均
- `SMA(S, N, M)` - 中国式移动平均
- `WMA(S, N)` - 加权移动平均
- `DMA(S, A)` - 动态移动平均

### 技术指标
- `MACD(CLOSE, SHORT, LONG, M)` - MACD指标
- `KDJ(CLOSE, HIGH, LOW, N, M1, M2)` - KDJ指标
- `RSI(CLOSE, N)` - RSI相对强弱指标
- `BOLL(CLOSE, N, P)` - 布林带指标
- `WR(CLOSE, HIGH, LOW, N, N1)` - 威廉指标
- `BIAS(CLOSE, L1, L2, L3)` - 乖离率
- `CCI(CLOSE, HIGH, LOW, N)` - 顺势指标
- `ATR(CLOSE, HIGH, LOW, N)` - 真实波动范围

### 高级指标
- `SAR(HIGH, LOW, N, S, M)` - 抛物转向指标
- `TDX_SAR(High, Low, iAFStep, iAFLimit)` - 通达信SAR算法
- `XSII(CLOSE, HIGH, LOW, N, M)` - 薛斯通道II

## 使用方法

### 基本用法

```go
package main

import (
    "fmt"
    "mylang/pkg/extensions/indicators"
)

func main() {
    // 创建价格序列
    close := indicators.NewSeries([]float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109})
    
    // 计算5日移动平均
    ma5 := indicators.MA(close, 5)
    fmt.Printf("MA5: %v\n", ma5)
    
    // 计算MACD指标
    dif, dea, macd := indicators.MACD(close, 12, 26, 9)
    fmt.Printf("MACD DIF: %v\n", dif)
    fmt.Printf("MACD DEA: %v\n", dea)
    fmt.Printf("MACD MACD: %v\n", macd)
}
```

### 使用指标函数映射

```go
// 获取指标函数
maFunc := indicators.GetIndicator("MA")
if maFunc != nil {
    result := maFunc([]any{close, 5})
    if result != nil {
        fmt.Printf("MA5: %v\n", result)
    }
}

// 注册自定义指标
indicators.RegisterIndicator("CUSTOM", func(args []any) any {
    // 自定义指标逻辑
    return nil
})
```

### 序列运算

```go
// 序列比较
greaterThan := indicators.GreaterThan(close, indicators.REF(close, 1))

// 统计函数
count := indicators.COUNT(greaterThan, 5)

// 数学运算
absClose := indicators.ABS(close)
```

## 数据结构

### Series类型
```go
type Series []float64
```

### 主要方法
- `NewSeries(data []float64) Series` - 创建新序列
- `Len() int` - 返回序列长度
- `At(i int) float64` - 获取指定位置元素
- `Last() float64` - 获取最后一个元素
- `Slice(start, end int) Series` - 获取子序列

## 注意事项

1. 所有指标函数都返回`Series`类型，包含NaN值表示无效数据
2. 序列长度不足时，函数会返回NaN值
3. 建议在实际使用前检查序列长度是否满足指标要求
4. 某些复杂指标可能需要更多历史数据才能得到准确结果

## 测试

运行测试用例：
```bash
go test -v
```

## 许可证

本项目基于原MyTT Python版本的许可证。
