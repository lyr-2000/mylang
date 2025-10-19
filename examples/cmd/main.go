package main

import (
	"fmt"
	"log"

	"github.com/lyr-2000/mylang/pkg/mylang"
)

func main() {
	// 创建一个新的麦语言解释器
	interp := mylang.NewMylangInterpreter()

	// 注册变量
	closeData := []float64{100.0, 101.0, 102.0, 103.0, 104.0}
	interp.RegisterVariable("CLOSE", closeData)
	fmt.Println("Registered CLOSE variable with data:", closeData)

	// 确认变量已注册
	if val, found := interp.GetVariable("CLOSE"); found {
		fmt.Println("Confirmed CLOSE variable is set to:", val)
	} else {
		fmt.Println("CLOSE variable not found in environment")
	}

	// 注册用户自定义函数
	interp.RegisterFunction("LLV", func(args []interface{}) interface{} {
		if len(args) != 2 {
			fmt.Println("LLV: Expected 2 arguments, got", len(args))
			return nil
		}
		data, ok1 := args[0].([]float64)
		n, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			fmt.Println("LLV: Invalid argument types, data:", args[0], "n:", args[1])
			return nil
		}
		if len(data) < int(n) {
			fmt.Println("LLV: Data length", len(data), "less than n", int(n))
			return nil
		}
		minVal := data[0]
		for i := 1; i < int(n); i++ {
			if data[i] < minVal {
				minVal = data[i]
			}
		}
		fmt.Println("LLV result:", minVal)
		return minVal
	})
	
	interp.RegisterVariable("NAMELIKE",func(args []interface{}) interface{} {
		log.Printf("Namelike args: %v", args)
		return false
	})

	interp.RegisterFunction("ADD", func(args []interface{}) interface{} {
		if len(args) != 2 {

			fmt.Println("HHV: Expected 2 arguments, got", len(args))
			return nil
		}
		data, ok1 := args[0].([]float64)
		n, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			fmt.Println("HHV: Invalid argument types, data:", args[0], "n:", args[1])
			return nil
		}
		for i := 0; i < len(data); i++ {
			data[i] += n
		}
		return data
	})

	// 执行麦语言代码 - 测试画图赋值
	code := `
S1:=IF(NAMELIKE('S'),0,1);
S2:=IF(NAMELIKE('*'),0,1);
S4:=IF(INBLOCK('科创板'),0,1);
S5:=IF(NAMELIKE('C'),0,1);
S6:=IF(INBLOCK('创业板'),0,1);
S7:=IF(INBLOCK('北证50'),0,1);
去除:=S1 AND S2 AND S4 AND S5 AND S6 AND S7;
a:=1
b:=2
c:a>b AND b=0
d:b=2
xx:HIGH>CLOSE
xy:HIGH<CLOSE
xz:HIGH=CLOSE
zy:HIGH!=CLOSE
zz:HIGH<=CLOSE
zzz:HIGH>=CLOSE
zzzz:HIGH==CLOSE
zzzzz:HIGH!=CLOSE
zzzzzz:HIGH<=CLOSE
zzzzzzz:HIGH>=CLOSE
    `
	result := interp.Execute(code)
	fmt.Println("Result:", result)

	// 调试输出
	fmt.Println("Debugging output:")
	if rsv, found := interp.GetVariable("RSV"); found {
		fmt.Printf("RSV: %v (Type: %T)\n", rsv, rsv)
	}
	if b, found := interp.GetVariable("B"); found {
		fmt.Printf("B: %v (Type: %T)\n", b, b)
	}
	if c, found := interp.GetVariable("C"); found {
		fmt.Printf("C: %v (Type: %T)\n", c, c)
	}

	// 检查画图变量
	fmt.Println("\nDrawing variables:")
	drawingVars := interp.GetDrawingVariables()
	if len(drawingVars) > 0 {
		for varName := range drawingVars {
			fmt.Printf("- %s (is drawing variable: %t)\n", varName, interp.IsDrawingVariable(varName))
			b,ok := interp.GetVariable(varName)
			fmt.Printf("%s : %v varok=%v\n",varName,b,ok)
		}
	} else {
		fmt.Println("No drawing variables found")
	}
}
