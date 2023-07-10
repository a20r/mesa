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

// Case represents a test case with its associated properties.
type Case[InstanceType, FieldsType, InputType, OutputType any] struct {
	Name       string                                                          // Name of the test case.
	Skip       string                                                          // Reason to skip the test case.
	Fields     FieldsType                                                      // Fields associated with the instance.
	Input      InputType                                                       // Input data for the test case.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)                 // Function to execute before calling the target function.
	Check      func(ctx *Ctx, inst InstanceType, in InputType, out OutputType) // Function to check the output of the target function.
	Cleanup    func(ctx *Ctx, inst InstanceType)                               // Cleanup function to execute after the test case finishes.
}

// Mesa represents a collection of test cases and the functions to create instances
// and execute the target function under test.
type Mesa[InstanceType, FieldsType, InputType, OutputType any] struct {
	NewInstance func(ctx *Ctx, fields FieldsType) InstanceType             // Function to create a new instance.
	Target      func(ctx *Ctx, inst InstanceType, in InputType) OutputType // Target function under test.
	Cases       []Case[InstanceType, FieldsType, InputType, OutputType]    // List of test cases.
}

// Run executes all the test cases in the Mesa instance.
func (m *Mesa[Inst, F, I, O]) Run(t *testing.T) {
	for _, tt := range m.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Skip != "" {
				t.Skip(tt.Skip)
			}

			re := require.New(t) // Create a new require.Assertions instance.
			as := assert.New(t)  // Create a new assert.Assertions instance.

			ctx := &Ctx{
				T:  t,
				As: as,
				Re: re,
			}

			inst := m.NewInstance(ctx, tt.Fields) // Create a new instance using NewInstance function.

			t.Cleanup(func() {
				tt.Cleanup(ctx, inst) // Execute the cleanup function after the test case finishes.
			})

			if tt.BeforeCall != nil {
				tt.BeforeCall(ctx, inst, tt.Input) // Execute the BeforeCall function before calling the target function.
			}

			out := m.Target(ctx, inst, tt.Input) // Call the target function with the provided input.

			if tt.Check != nil {
				tt.Check(ctx, inst, tt.Input, out) // Check the output of the target function using the Check function.
			}
		})
	}
}
