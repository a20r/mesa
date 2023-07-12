# Mesa
[![Go](https://github.com/a20r/mesa/actions/workflows/go.yml/badge.svg)](https://github.com/a20r/mesa/actions/workflows/go.yml)
[![golangci-lint](https://github.com/a20r/mesa/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/a20r/mesa/actions/workflows/golangci-lint.yml)

Mesa is a package for creating and running table driven tests

# Install
```
go get github.com/a20r/mesa
```

# Example

## Testing a method
```go
package buffer_test

import (
    "fmt"
    "testing"

    "github.com/a20r/mesa"
)

type Msg struct {
    Name  string
    Value int
}

type Buffer struct {
    msgs  []Msg
    limit int
}

var ErrBufferIsFull = fmt.Errorf("cannot add item: buffer full")

func (a *Buffer) Add(msg Msg) error {
    if len(a.msgs) >= a.limit {
        return ErrBufferIsFull
    }

    a.msgs = append(a.msgs, msg)
    return nil
}

func TestBuffer_Add(t *testing.T) {
    m := mesa.InstanceMesa[*Buffer, int, Msg, error]{
        NewInstance: func(ctx *mesa.Ctx, limit int) *Buffer {
            return &Buffer{
                limit: limit,
            }
        },
        Target: func(ctx *mesa.Ctx, inst *Buffer, in Msg) error {
            return inst.Add(in)
        },
        Cases: []mesa.InstanceCase[*Buffer, int, Msg, error]{
            {
                Name:   "Buffer with limit 10",
                Fields: 10,
                Input: Msg{
                    Name:  "test-value",
                    Value: 42,
                },
                Check: func(ctx *mesa.Ctx, inst *Buffer, in Msg, out error) {
                    if ctx.As.NoError(out) && ctx.As.Len(inst.msgs, 1) {
                        ctx.As.Equal(in, inst.msgs[0])
                    }
                },
            },
            {
                Name:   "Buffer is full",
                Fields: 10,
                Input: Msg{
                    Name:  "test-value",
                    Value: 42,
                },
                BeforeCall: func(ctx *mesa.Ctx, inst *Buffer, in Msg) {
                    for i := 0; i < inst.limit; i++ {
                        ctx.Re.NoError(inst.Add(in))
                    }
                },
                Check: func(ctx *mesa.Ctx, inst *Buffer, in Msg, out error) {
                    if ctx.As.Error(out) {
                        ctx.As.ErrorIs(out, ErrBufferIsFull)
                    }
                },
            },
            {
                Name:   "Test is skipped",
                Skip:   "Skipping test because it fails for now",
                Fields: -1,
                Input: Msg{
                    Name:  "test-value",
                    Value: 42,
                },
            },
        },
    }

    m.Run(t)
}
```

## Testing a function
```go
package pow_test

import (
    "math"
    "testing"

    "github.com/a20r/mesa"
)

func PowE(x float64) float64 {
    return math.Pow(math.E, x)
}

func TestPowE(t *testing.T) {
    m := mesa.FunctionMesa[float64, float64]{
        Target: func(ctx *mesa.Ctx, in float64) float64 {
            return PowE(in)
        },
        Cases: []mesa.FunctionCase[float64, float64]{
            {
                Name:  "pow(e, 1)",
                Input: 1,
                Check: func(ctx *mesa.Ctx, in, out float64) {
                    ctx.As.InEpsilon(math.E, out, 0.00001)
                },
            },
        },
    }

    m.Run(t)
}

```
