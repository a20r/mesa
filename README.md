# Mesa
[![Go](https://github.com/a20r/mesa/actions/workflows/go.yml/badge.svg)](https://github.com/a20r/mesa/actions/workflows/go.yml)
[![golangci-lint](https://github.com/a20r/mesa/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/a20r/mesa/actions/workflows/golangci-lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/a20r/mesa.svg)](https://pkg.go.dev/github.com/a20r/mesa)

Mesa is a package for creating and running table driven tests in Go.

# Install
```
go get github.com/a20r/mesa
```

# Usage

Mesa provides two types of testing: method testing and function testing.

## Testing methods

Method testing is used to test methods of a struct. To use Mesa for method testing, create a `MethodMesa` instance and define the following:

- `NewInstance`: a function that creates a new instance of the struct being tested
- `Target`: the method being tested
- `Cases`: an array of `MethodCase` instances that define the test cases
- `BeforeCall`: an optional function to execute before calling the target method
- `Check`: an optional function to check the output of the target method
- `Cleanup`: an optional function to execute after the test case finishes


Each `MethodCase` instance defines the following:

- `Name`: the name of the test case
- `Fields` or `FieldsFn`: the fields of the struct being tested
- `Input` or `InputFn`: the input to the method being tested
- `Skip`: an optional reason to skip the test case
- [*Override*] `BeforeCall`: an optional function to execute before calling the target method
- [*Override*] `Check`: an optional function to check the output of the target method
- [*Override*] `Cleanup`: an optional function to execute after the test case finishes

### Example
```go
type MyStruct struct {
    Value int
}

func (s *MyStruct) Add(n int) {
    s.Value += n
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
```

## Testing functions
Function testing is used to test standalone functions. To use Mesa for function testing, create a `FunctionMesa` instance and define the following:

- `Target`: the function being tested
- `Cases`: an array of `FunctionCase` instances that define the test cases
- `BeforeCall`: an optional function to execute before calling the target function
- `Check`: an optional function to check the output of the target function
- `Cleanup`: an optional function to execute after the test case finishes

Each `FunctionCase` instance defines the following:

- `Name`: the name of the test case
- `Input` or `InputFn`: the input to the function being tested
- `Skip`: an optional reason to skip the test case
- `Check`: an optional function to check the output of the target function
- [*Override*] `BeforeCall`: an optional function to execute before calling the target function
- [*Override*] `Check`: an optional function to check the output of the target function
- [*Override*] `Cleanup`: an optional function to execute after the test case finishes

### Example
```go
func Add(a, b int) int {
    return a + b
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
```

# Contributing

Contributions are welcome! Please see the [contributing guidelines](CONTRIBUTING.md) for more information.

# License

Mesa is licensed under the [MIT License](LICENSE).
