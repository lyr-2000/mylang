package api

import (
	"fmt"
	"io"
	"testing"
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

	x.SetVar("CLOSE", []float64{1, 2, 3, 4, 5})

	err := x.RunCode("d:=ADD(CLOSE,5)")
	if err != nil {
		t.Errorf("RunCode error: %v", err)
	}
	fmt.Println(x.GetVariable("d"))

}
