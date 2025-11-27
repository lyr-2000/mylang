# Mylang 解释器

本项目实现了一个简单的（Mylang）解释器，支持基本的变量操作、表达式求值、数组运算以及用户自定义函数。整个解释器采用 Go 语言实现，便于进行二次开发或嵌入到更大的系统中。

## 特性

- **变量管理**：支持设置和获取变量。
- **四则运算**：支持浮点数和数组的加减乘除，以及它们之间的混合运算（例如数组与浮点数、数组与数组等）。
- **自定义函数**：可以注册自定义 Go 函数到解释器环境，用于后续调用。
- **错误处理**：函数调用等环节有基础的错误处理和日志输出，便于调试。

## 文件结构说明

- `pkg/mylang/interpreter.go`：主解释器功能，包括环境、表达式求值、函数调用等。
- `pkg/mylang/api.go`：对外接口，包含变量和函数注册、执行接口等。

## 快速开始

1. **初始化解释器：**

```go
mi := mylang.NewMylangInterpreter()
```

2. **注册变量和自定义函数：**

```go
package api

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestMaiExecutor_RunCode(t *testing.T) {
	SetOutput(io.Discard)
	x := NewMaiExecutor()
	x.RegisterFunction("ADD", func(args []interface{}) interface{} {
		d := Arrayfloat64(args[0])
		for i, v := range d {
			d[i] = v + args[1].(float64)
		}
		return d
	})


	x.SetCustomVariableGetter(func(name string) any {
		if name == "NOWDAY" {
			return time.Now().Format("20060102")
		}
		return nil
	})
	x.SetVar("CLOSE", []float64{1, 2, 3, 4, 5})

	err := x.RunCode("d:=ADD(CLOSE,5)\nb:=NOWDAY")
	if err != nil {
		t.Errorf("RunCode error: %v", err)
	}
	fmt.Println(x.GetVariable("d"))
	fmt.Println(x.GetVariable("b"))

}


```


