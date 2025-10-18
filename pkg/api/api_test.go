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
