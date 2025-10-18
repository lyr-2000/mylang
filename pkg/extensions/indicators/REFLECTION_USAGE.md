# MyTT Go版本 - 反射调用使用指南

## 概述

MyTT Go版本提供了真正的反射调用功能，使用Go的`reflect`包来动态调用指标函数。这意味着你可以传入`[]any`类型的参数，系统会自动识别函数签名并进行类型转换。

## 核心功能

### 1. CallIndicatorByReflection - 反射调用函数

```go
func CallIndicatorByReflection(funcName string, args []any) (any, error)
```

这是主要的反射调用函数，它会：
- 根据函数名获取原始函数
- 使用反射检查参数类型
- 自动进行类型转换
- 调用函数并返回结果

### 2. 支持的函数

所有MyTT库中的函数都支持反射调用，包括：

#### 0级核心工具函数
- `RD`, `RET`, `ABS`, `LN`, `POW`, `SQRT`
- `SIN`, `COS`, `TAN`, `MAX`, `MIN`
- `REF`, `DIFF`, `STD`, `SUM`
- `MA`, `EMA`, `SMA`, `WMA`, `DMA`
- `HHV`, `LLV`, `AVEDEV`, `SLOPE`, `FORCAST`

#### 1级应用层函数
- `COUNT`, `EVERY`, `EXIST`, `CROSS`, `LONGCROSS`
- `BARSLAST`, `VALUEWHEN`, `BETWEEN`, `FILTER`, `LAST`
- `BARSSINCEN`, `HHVBARS`, `LLVBARS`, `TOPRANGE`, `LOWRANGE`

#### 2级技术指标函数
- `MACD`, `RSI`, `BOLL`, `KDJ`, `WR`, `BIAS`
- `PSY`, `CCI`, `ATR`, `BBI`, `DMI`, `TAQ`
- `KTN`, `TRIX`, `VR`, `CR`, `EMV`, `DPO`
- `BRAR`, `DFMA`, `MTM`, `MASS`, `ROC`
- `EXPMA`, `OBV`, `MFI`, `ASI`, `XSII`

#### 高级函数版本
- `SAR`, `TDX_SAR`, `QRR`, `SHO`, `LON`

## 使用示例

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
    
    // 示例1: SMA(close, 100, 1.0) - 你提到的例子
    result, err := indicators.CallIndicatorByReflection("SMA", []any{close, 100, 1.0})
    if err != nil {
        fmt.Printf("错误: %v\n", err)
    } else {
        fmt.Printf("SMA结果: %v\n", result)
    }
    
    // 示例2: MA(close, 5)
    result, err = indicators.CallIndicatorByReflection("MA", []any{close, 5})
    if err != nil {
        fmt.Printf("错误: %v\n", err)
    } else {
        fmt.Printf("MA结果: %v\n", result)
    }
    
    // 示例3: ABS(close)
    result, err = indicators.CallIndicatorByReflection("ABS", []any{close})
    if err != nil {
        fmt.Printf("错误: %v\n", err)
    } else {
        fmt.Printf("ABS结果: %v\n", result)
    }
}
```

### 多返回值函数

```go
// MACD函数返回三个值
result, err := indicators.CallIndicatorByReflection("MACD", []any{close, 12, 26, 9})
if err != nil {
    fmt.Printf("错误: %v\n", err)
} else {
    // result是一个切片，包含三个Series
    results := result.([]any)
    dif := results[0].(indicators.Series)
    dea := results[1].(indicators.Series)
    macd := results[2].(indicators.Series)
    fmt.Printf("DIF: %v\n", dif)
    fmt.Printf("DEA: %v\n", dea)
    fmt.Printf("MACD: %v\n", macd)
}
```

### 参数类型转换

反射系统会自动处理类型转换：

```go
// 这些调用都是等价的
result1, _ := indicators.CallIndicatorByReflection("SMA", []any{close, 5, 1.0})
result2, _ := indicators.CallIndicatorByReflection("SMA", []any{close, int(5), float64(1)})
result3, _ := indicators.CallIndicatorByReflection("MA", []any{close, 5})
result4, _ := indicators.CallIndicatorByReflection("MA", []any{close, int(5)})
```

### 错误处理

```go
// 不存在的函数
result, err := indicators.CallIndicatorByReflection("NOTEXIST", []any{close, 5})
if err != nil {
    fmt.Printf("预期的错误: %v\n", err) // original function for 'NOTEXIST' not found
}

// 参数数量不匹配
result, err = indicators.CallIndicatorByReflection("MA", []any{close})
if err != nil {
    fmt.Printf("预期的错误: %v\n", err) // argument count mismatch: expected 2, got 1
}

// 无效的参数类型
result, err = indicators.CallIndicatorByReflection("MA", []any{close, "invalid"})
if err != nil {
    fmt.Printf("预期的错误: %v\n", err) // cannot convert argument 1 from string to int
}
```

## 技术实现

### 反射调用流程

1. **函数查找**: 根据函数名在`getOriginalFunction`中查找原始函数
2. **类型检查**: 使用`reflect.ValueOf`和`reflect.Type`检查函数签名
3. **参数转换**: 自动转换参数类型以匹配函数签名
4. **函数调用**: 使用`reflect.Value.Call`调用函数
5. **结果处理**: 处理单个或多个返回值

### 类型转换规则

- `int` 可以转换为 `float64`
- `float64` 可以转换为 `int`（如果值在整数范围内）
- `Series` 类型必须完全匹配
- 其他类型转换遵循Go的类型转换规则

### 性能考虑

- 反射调用比直接调用慢，但在动态调用场景下是必要的
- 函数查找使用switch语句，性能较好
- 参数类型转换只在必要时进行

## 扩展功能

### 添加新函数

要添加新的指标函数到反射系统中：

1. 实现函数
2. 在`getOriginalFunction`中添加case
3. 函数会自动支持反射调用

```go
// 1. 实现函数
func MyCustomIndicator(data Series, param int) Series {
    // 实现逻辑
    return result
}

// 2. 在getOriginalFunction中添加
case "MyCustomIndicator":
    return MyCustomIndicator

// 3. 现在可以反射调用
result, err := CallIndicatorByReflection("MyCustomIndicator", []any{data, 10})
```

## 总结

MyTT Go版本的反射调用功能提供了：

- ✅ 真正的反射调用，使用Go的`reflect`包
- ✅ 自动类型转换和参数验证
- ✅ 支持所有指标函数
- ✅ 完整的错误处理
- ✅ 多返回值支持
- ✅ 高性能的函数查找

这使得MyTT库可以用于动态指标计算场景，如公式解析器、策略引擎等。
