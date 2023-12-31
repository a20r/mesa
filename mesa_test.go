package mesa_test

import (
	"testing"

	"github.com/a20r/mesa"
)

func BenchmarkTest(b *testing.B) {
	m := mesa.MethodBenchmarkMesa[*MyStruct, int, int, mesa.Empty]{
		NewInstance: func(ctx *mesa.Ctx, value int) *MyStruct {
			return &MyStruct{Value: value}
		},
		Target: func(ctx *mesa.Ctx, inst *MyStruct, n int) mesa.Empty {
			inst.Add(n)
			return nil
		},
		Cases: []mesa.MethodBenchmarkCase[*MyStruct, int, int, mesa.Empty]{
			{
				Name:   "Add 1 to 0",
				Fields: 0,
				Input:  1,
			},
			{
				Name:   "Add 2 to 1",
				Fields: 1,
				Input:  2,
			},
		},
	}

	m.Run(b)
}
