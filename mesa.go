// Mesa is a package for creating and running table driven tests
package mesa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// InstanceCase represents a test case with its associated properties.
type InstanceCase[InstanceType, FieldsType, InputType, OutputType any] struct {
	Name       string                                                          // Name of the test case.
	Skip       string                                                          // Reason to skip the test case.
	Fields     FieldsType                                                      // Fields associated with the instance.
	Input      InputType                                                       // Input data for the test case.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)                 // Function to execute before calling the target function.
	Check      func(ctx *Ctx, inst InstanceType, in InputType, out OutputType) // Function to check the output of the target function.
	Cleanup    func(ctx *Ctx, inst InstanceType)                               // Cleanup function to execute after the test case finishes.
}

// InstanceMesa represents a collection of test cases and the functions to create instances
// and execute the target function under test.
type InstanceMesa[InstanceType, FieldsType, InputType, OutputType any] struct {
	NewInstance func(ctx *Ctx, fields FieldsType) InstanceType                  // Function to create a new instance.
	Target      func(ctx *Ctx, inst InstanceType, in InputType) OutputType      // Target function under test.
	Cases       []InstanceCase[InstanceType, FieldsType, InputType, OutputType] // List of test cases.
}

// Run executes all the test cases in the Mesa instance.
func (m *InstanceMesa[Inst, F, I, O]) Run(t *testing.T) {
	for _, tt := range m.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Skip != "" {
				t.Skip(tt.Skip)
			}

			ctx := newCtx(t)

			inst := m.NewInstance(ctx, tt.Fields)

			if tt.Cleanup != nil {
				t.Cleanup(func() {
					tt.Cleanup(ctx, inst)
				})
			}

			if tt.BeforeCall != nil {
				tt.BeforeCall(ctx, inst, tt.Input)
			}

			out := m.Target(ctx, inst, tt.Input)

			if tt.Check != nil {
				tt.Check(ctx, inst, tt.Input, out)
			}
		})
	}
}

// FunctionCase represents a test case with its associated properties.
type FunctionCase[InputType, OutputType any] struct {
	Name       string                                       // Name of the test case.
	Skip       string                                       // Reason to skip the test case.
	Input      InputType                                    // Input data for the test case.
	BeforeCall func(ctx *Ctx, in InputType)                 // Function to execute before calling the target function.
	Check      func(ctx *Ctx, in InputType, out OutputType) // Function to check the output of the target function.
	Cleanup    func(ctx *Ctx)                               // Cleanup function to execute after the test case finishes.
}

// FunctionMesa represents a collection of test cases executes the target function under each test case.
type FunctionMesa[InputType, OutputType any] struct {
	Target func(ctx *Ctx, in InputType) OutputType
	Cases  []FunctionCase[InputType, OutputType]
}

// Run executes all the test cases in the FunctionMesa instance.
func (m *FunctionMesa[I, O]) Run(t *testing.T) {
	for _, tt := range m.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Skip != "" {
				t.Skip(tt.Skip)
			}

			ctx := newCtx(t)

			if tt.Cleanup != nil {
				t.Cleanup(func() {
					tt.Cleanup(ctx)
				})
			}

			if tt.BeforeCall != nil {
				tt.BeforeCall(ctx, tt.Input)
			}

			out := m.Target(ctx, tt.Input)

			if tt.Check != nil {
				tt.Check(ctx, tt.Input, out)
			}
		})
	}
}
