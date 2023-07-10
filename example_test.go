package mesa_test

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

func ExampleMesa() {
	m := mesa.Mesa[*Buffer, int, Msg, error]{
		NewInstance: func(ctx *mesa.Ctx, limit int) *Buffer {
			return &Buffer{
				limit: limit,
			}
		},
		Target: func(ctx *mesa.Ctx, inst *Buffer, in Msg) error {
			return inst.Add(in)
		},
		Cases: []mesa.Case[*Buffer, int, Msg, error]{
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

	// Dummy testing variable
	var t = &testing.T{}

	m.Run(t)
}
