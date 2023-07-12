package mesa_test

import (
	"testing"

	"github.com/a20r/mesa"
)

type MyStruct struct {
	Value int
}

func (s *MyStruct) Add(n int) {
	s.Value += n
}

func ExampleMethodMesa() {
	//   type MyStruct struct {
	//     Value int
	//   }
	//
	//   func (s *MyStruct) Add(n int) {
	// 	   s.Value += n
	//   }

	m := mesa.MethodMesa[*MyStruct, int, int, mesa.Empty]{
		NewInstance: func(ctx *mesa.Ctx, value int) *MyStruct {
			return &MyStruct{Value: value}
		},
		Target: func(ctx *mesa.Ctx, inst *MyStruct, n int) mesa.Empty {
			inst.Add(n)
			return nil
		},
		Cases: []mesa.MethodCase[*MyStruct, int, int, mesa.Empty]{
			{
				Name:   "Add 1 to 0",
				Fields: 0,
				Input:  1,
				Check: func(ctx *mesa.Ctx, inst *MyStruct, in int, _ mesa.Empty) {
					ctx.As.Equal(1, inst.Value)
				},
			},
			{
				Name:   "Add 2 to 1",
				Fields: 1,
				Input:  2,
				Check: func(ctx *mesa.Ctx, inst *MyStruct, in int, _ mesa.Empty) {
					ctx.As.Equal(3, inst.Value)
				},
			},
		},
	}

	var t *testing.T
	m.Run(t)
}

func TestMyStruct_Add(t *testing.T) {
	m := mesa.MethodMesa[*MyStruct, int, int, mesa.Empty]{
		NewInstance: func(ctx *mesa.Ctx, value int) *MyStruct {
			return &MyStruct{Value: value}
		},
		Target: func(ctx *mesa.Ctx, inst *MyStruct, n int) mesa.Empty {
			inst.Add(n)
			return nil
		},
		Cases: []mesa.MethodCase[*MyStruct, int, int, mesa.Empty]{
			{
				Name:   "Add 1 to 0",
				Fields: 0,
				Input:  1,
				Check: func(ctx *mesa.Ctx, inst *MyStruct, in int, _ mesa.Empty) {
					ctx.As.Equal(1, inst.Value)
				},
			},
			{
				Name:   "Add 2 to 1",
				Fields: 1,
				Input:  2,
				Check: func(ctx *mesa.Ctx, inst *MyStruct, in int, _ mesa.Empty) {
					ctx.As.Equal(3, inst.Value)
				},
			},
		},
	}

	m.Run(t)
}

func Add(a, b int) int {
	return a + b
}

func ExampleFunctionMesa() {
	//   func Add(a, b int) int {
	// 	   return a + b
	//   }

	type input struct{ a, b int }

	m := mesa.FunctionMesa[input, int]{
		Target: func(ctx *mesa.Ctx, in input) int {
			return Add(in.a, in.b)
		},
		Cases: []mesa.FunctionCase[input, int]{
			{
				Name:  "Add 1 and 2",
				Input: input{a: 1, b: 2},
				Check: func(ctx *mesa.Ctx, in input, out int) {
					ctx.As.Equal(3, out)
				},
			},
			{
				Name:  "Add 0 and 0",
				Input: input{a: 0, b: 0},
				Check: func(ctx *mesa.Ctx, in input, out int) {
					ctx.As.Equal(0, out)
				},
			},
		},
	}

	var t *testing.T
	m.Run(t)
}

func TestAdd(t *testing.T) {
	type input struct{ a, b int }

	m := mesa.FunctionMesa[input, int]{
		Target: func(ctx *mesa.Ctx, in input) int {
			return Add(in.a, in.b)
		},
		Cases: []mesa.FunctionCase[input, int]{
			{
				Name:  "Add 1 and 2",
				Input: input{a: 1, b: 2},
				Check: func(ctx *mesa.Ctx, in input, out int) {
					ctx.As.Equal(3, out)
				},
			},
			{
				Name:  "Add 0 and 0",
				Input: input{a: 0, b: 0},
				Check: func(ctx *mesa.Ctx, in input, out int) {
					ctx.As.Equal(0, out)
				},
			},
		},
	}

	m.Run(t)
}
