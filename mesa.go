// Mesa is a package for creating and running table driven tests
package mesa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mesa is an interface that defines a method to run a test suite.
type Mesa interface {
	// Run runs the test suite.
	Run(t *testing.T)
}

// Run runs the provided test suites.
func Run(t *testing.T, ms ...Mesa) {
	for _, m := range ms {
		m.Run(t)
	}
}

// Ctx represents the test context containing the testing.T instance
// and assertion objects for convenience.
type Ctx struct {
	T  *testing.T
	As *assert.Assertions
	Re *require.Assertions
}

// newCtx creates a new testing context with assert and require instances.
func newCtx(t *testing.T) *Ctx {
	return &Ctx{
		T:  t,
		As: assert.New(t),
		Re: require.New(t),
	}
}

// MethodCase represents a test case with its associated properties.
type MethodCase[InstanceType, FieldsType, InputType, OutputType any] struct {
	// Required: Name of the test case.
	Name string

	// Required: Fields associated with the instance.
	Fields FieldsType

	// Required: Input data for the test case.
	Input InputType

	// Optional: Reason to skip the test case. The test is only skipped if this field is not empty
	Skip string

	// Optional: Function to execute before calling the target function. It will be called instead of the BeforeCall
	// function in the MethodMesa if provided.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// Optional: Function to check the output of the target function. It will be called instead of the Check function
	// in the MethodMesa if provided.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// Optional: Cleanup function to execute after the test case finishes. It will be called instead of the Cleanup
	// function in the MethodMesa if provided.
	Cleanup func(ctx *Ctx, inst InstanceType)
}

// MethodMesa represents a collection of test cases and the functions to create instances
// and execute the target function under test.
type MethodMesa[InstanceType, FieldsType, InputType, OutputType any] struct {
	// Required: Function to create a new instance.
	NewInstance func(ctx *Ctx, fields FieldsType) InstanceType

	// Required: Target function under test.
	Target func(ctx *Ctx, inst InstanceType, in InputType) OutputType

	// Required: List of test cases.
	Cases []MethodCase[InstanceType, FieldsType, InputType, OutputType]

	// Optional: Function to execute before calling the target function. This is called when no BeforeCall function
	// is provided by the the case itself.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// Optional: Function to check the output of the target function. This is called when no Check function
	// is provided by the the case itself.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// Optional: Cleanup function to execute after the test case finishes. This is called when no Cleanup function
	// is provided by the the case itself.
	Cleanup func(ctx *Ctx, inst InstanceType)
}

// Run executes all the test cases in the Mesa instance.
func (m *MethodMesa[Inst, F, I, O]) Run(t *testing.T) {
	for _, tt := range m.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Skip != "" {
				t.Skip(tt.Skip)
			}

			ctx := newCtx(t)

			inst := m.NewInstance(ctx, tt.Fields)

			cleanup := func() {}

			switch {
			case tt.Cleanup != nil:
				cleanup = func() { tt.Cleanup(ctx, inst) }
			case m.Cleanup != nil:
				cleanup = func() { m.Cleanup(ctx, inst) }
			}

			t.Cleanup(cleanup)

			switch {
			case tt.BeforeCall != nil:
				tt.BeforeCall(ctx, inst, tt.Input)
			case m.BeforeCall != nil:
				m.BeforeCall(ctx, inst, tt.Input)
			}

			out := m.Target(ctx, inst, tt.Input)

			switch {
			case tt.Check != nil:
				tt.Check(ctx, inst, tt.Input, out)
			case m.Check != nil:
				m.Check(ctx, inst, tt.Input, out)
			}
		})
	}
}

// FunctionCase represents a test case with its associated properties.
type FunctionCase[InputType, OutputType any] struct {
	// Required: Name of the test case.
	Name string

	// Required: Input data for the test case.
	Input InputType

	// Optional: Reason to skip the test case. The test is only skipped if this field is not empty
	Skip string

	// Optional: Function to execute before calling the target function. It will be called instead of the BeforeCall
	// function in the FunctionMesa if provided.
	BeforeCall func(ctx *Ctx, in InputType)

	// Optional: Function to check the output of the target function. It will be called instead of the Check
	// function in the FunctionMesa if provided.
	Check func(ctx *Ctx, in InputType, out OutputType)

	// Optional: Cleanup function to execute after the test case finishes. It will be called instead of the Cleanup
	// function in the FunctionMesa if provided.
	Cleanup func(ctx *Ctx)
}

// FunctionMesa represents a collection of test cases that execute the target function under each test case.
type FunctionMesa[InputType, OutputType any] struct {
	// Required: Target function under test.
	Target func(ctx *Ctx, in InputType) OutputType

	// Required: List of test cases.
	Cases []FunctionCase[InputType, OutputType]

	// Optional: Function to execute before calling the target function. This is called when no BeforeCall function
	// is provided by the the case itself.
	BeforeCall func(ctx *Ctx, in InputType)

	// Optional: Function to check the output of the target function. This is called when no Check function
	// is provided by the the case itself.
	Check func(ctx *Ctx, in InputType, out OutputType)

	// Optional: Cleanup function to execute after the test case finishes. This is called when no Cleanup function
	// is provided by the the case itself.
	Cleanup func(ctx *Ctx)
}

// Run executes all the test cases in the FunctionMesa instance.
func (m *FunctionMesa[I, O]) Run(t *testing.T) {
	im := MethodMesa[any, any, I, O]{
		NewInstance: func(_ *Ctx, _ any) any {
			return nil
		},

		Target: func(ctx *Ctx, _ any, in I) O {
			return m.Target(ctx, in)
		},

		BeforeCall: func(ctx *Ctx, _ any, in I) {
			if m.BeforeCall != nil {
				m.BeforeCall(ctx, in)
			}
		},

		Check: func(ctx *Ctx, _ any, in I, out O) {
			if m.Check != nil {
				m.Check(ctx, in, out)
			}
		},

		Cleanup: func(ctx *Ctx, _ any) {
			if m.Cleanup != nil {
				m.Cleanup(ctx)
			}
		},

		Cases: make([]MethodCase[any, any, I, O], len(m.Cases)),
	}

	for i, c := range m.Cases {
		im.Cases[i] = MethodCase[any, any, I, O]{
			Name:  c.Name,
			Input: c.Input,
			Skip:  c.Skip,

			BeforeCall: func(ctx *Ctx, _ any, in I) {
				if c.BeforeCall != nil {
					c.BeforeCall(ctx, in)
				}
			},

			Check: func(ctx *Ctx, _ any, in I, out O) {
				if c.Check != nil {
					c.Check(ctx, in, out)
				}
			},

			Cleanup: func(ctx *Ctx, _ any) {
				if c.Cleanup != nil {
					c.Cleanup(ctx)
				}
			},
		}
	}

	im.Run(t)
}
