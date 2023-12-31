// Mesa is a package for creating and running table driven tests
package mesa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mesa is an interface that defines a method to run a test suite.
type Mesa interface {
	// Run runs the test suite.
	Run(t *testing.T)
}

// Type assertions to ensure Method and Function mesas adhere to interface
var (
	_ Mesa = MethodMesa[any, any, any, any]{}
	_ Mesa = FunctionMesa[any, any]{}
)

// Run runs the provided test suites.
func Run(t *testing.T, ms ...Mesa) {
	for _, m := range ms {
		m.Run(t)
	}
}

// Empty is a type used when testing structs and functions without fields or return values
type Empty any

// ErrorPair is a convenience type used to wrap function outputs that return a value and an error
type ErrorPair[T any] struct {
	Value T
	Err   error
}

// NewErrorPair creates a new error pair with the provided value and error
func NewErrorPair[T any](value T, err error) ErrorPair[T] {
	return ErrorPair[T]{Value: value, Err: err}
}

// Ctx represents the test context containing the testing.T instance
// and assertion objects for convenience.
type Ctx struct {
	context.Context
	t       require.TestingT
	values  map[string]any
	metrics map[string]float64
	As      *assert.Assertions
	Re      *require.Assertions
}

// T returns the underlying testing.T instance if it is being used tests. The test will fail if the Ctx is being
// used for benchmarking.
func (c *Ctx) T() *testing.T {
	t, ok := c.t.(*testing.T)
	c.Re.True(ok, "Ctx is being used for a benchmark but you are trying to retrieve the *testing.T instance")
	return t
}

// B returns the underlying testing.B instance if it is being used for benchmarking. The benchmark will fail if
// the Ctx is being used for tests.
func (c *Ctx) B() *testing.B {
	b, ok := c.t.(*testing.B)
	c.Re.True(ok, "Ctx is being used for a test but you are trying to retrieve the *testing.B instance")
	return b
}

// ReportMetric adds the benchmarking metric with the given name to the context. The metrics are divided by the number
// of calls (b.N) once the benchmark is complete. It will panic if the context is being used for tests
func (c *Ctx) ReportMetric(value float64, name string) {
	b := c.B()
	b.StopTimer()
	defer b.StartTimer()

	c.metrics[name] += value
}

// SetValue sets a value with the given name in the context
func (c *Ctx) SetValue(name string, val any) {
	c.values[name] = val
}

// GetValue gets a value with the given name from the context
func (c *Ctx) GetValue(name string) any {
	return c.values[name]
}

// newCtx creates a new testing context with assert and require instances.
func newCtx(t require.TestingT) *Ctx {
	return &Ctx{
		Context: context.Background(),
		t:       t,
		values:  make(map[string]any),
		metrics: make(map[string]float64),
		As:      assert.New(t),
		Re:      require.New(t),
	}
}

// MethodCase represents a test case with its associated properties.
type MethodCase[InstanceType, FieldsType, InputType, OutputType any] struct {
	// [Required] Name of the test case.
	Name string

	// [Optional] Fields associated with the instance. FieldsFn takes priority over Fields. If fields are not needed
	// to instantiate a the test instance, no fields need to be provided.
	Fields FieldsType

	// [Optional] FieldsFn returns the fields used for this case. FieldsFn takes priority over Fields. If fields are
	// not needed to instantiate a the test instance, no fields need to be provided.
	FieldsFn func(ctx *Ctx) FieldsType

	// [Optional] Input data for the test case. InputFn takes priority over Input. The Input field can be empty if the
	// target function does not take any arguments.
	Input InputType

	// [Optional] InputFn returns the input struct used for this case. It takes priority over the Input field. This can
	// be empty if the target function does not take any arguments.
	InputFn func(ctx *Ctx, inst InstanceType) InputType

	// [Optional] Reason to skip the test case. The test is only skipped if this field is not empty
	Skip string

	// [Optional] Function to execute before calling the target function. It will be called instead of the BeforeCall
	// function in the MethodMesa if provided.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// [Optional] Function to check the output of the target function. It will be called instead of the Check function
	// in the MethodMesa if provided.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// [Optional] Cleanup function to execute after the test case finishes. It will be called instead of the Cleanup
	// function in the MethodMesa if provided.
	Cleanup func(ctx *Ctx, inst InstanceType)
}

// MethodMesa represents a collection of test cases and the functions to create instances
// and execute the target function under test.
type MethodMesa[InstanceType, FieldsType, InputType, OutputType any] struct {
	// [Optional] Function to initialize anything before running the test cases
	Init func(ctx *Ctx)

	// [Required] Function to create a new instance.
	NewInstance func(ctx *Ctx, fields FieldsType) InstanceType

	// [Required] Target function under test.
	Target func(ctx *Ctx, inst InstanceType, in InputType) OutputType

	// [Required] List of test cases.
	Cases []MethodCase[InstanceType, FieldsType, InputType, OutputType]

	// [Optional] Function to execute before calling the target function. This is called when no BeforeCall function
	// is provided by the the case itself.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// [Optional] Function to check the output of the target function. This is called when no Check function
	// is provided by the the case itself.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// [Optional] Cleanup function to execute after the test case finishes. This is called when no Cleanup function
	// is provided by the the case itself.
	Cleanup func(ctx *Ctx, inst InstanceType)

	// [Optional] Teardown function is called after all cases finish
	Teardown func(ctx *Ctx)
}

// Run executes all the test cases in the Mesa instance.
func (m MethodMesa[Inst, F, I, O]) Run(t *testing.T) {
	ctx := newCtx(t)

	if m.Init != nil {
		m.Init(ctx)
	}

	if m.Teardown != nil {
		defer m.Teardown(ctx)
	}

	for _, tt := range m.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Skip != "" {
				t.Skip(tt.Skip)
			}

			ctx := newCtx(t)

			if tt.FieldsFn != nil {
				tt.Fields = tt.FieldsFn(ctx)
			}

			inst := m.NewInstance(ctx, tt.Fields)

			if tt.InputFn != nil {
				tt.Input = tt.InputFn(ctx, inst)
			}

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
	// [Required] Name of the test case.
	Name string

	// [Optional] Input data for the test case. InputFn takes priority over Input. The Input field can be empty if the
	// target function does not take any arguments.
	Input InputType

	// [Optional] InputFn returns the input struct used for this case. It takes priority over the Input field. This can
	// be empty if the target function does not take any arguments.
	InputFn func(ctx *Ctx) InputType

	// [Optional] Reason to skip the test case. The test is only skipped if this field is not empty
	Skip string

	// [Optional] Function to execute before calling the target function. It will be called instead of the BeforeCall
	// function in the FunctionMesa if provided.
	BeforeCall func(ctx *Ctx, in InputType)

	// [Optional] Function to check the output of the target function. It will be called instead of the Check
	// function in the FunctionMesa if provided.
	Check func(ctx *Ctx, in InputType, out OutputType)

	// [Optional] Cleanup function to execute after the test case finishes. It will be called instead of the Cleanup
	// function in the FunctionMesa if provided.
	Cleanup func(ctx *Ctx)
}

// FunctionMesa represents a collection of test cases that execute the target function under each test case.
type FunctionMesa[InputType, OutputType any] struct {
	// [Optional] Function to initialize anything before running the test cases
	Init func(ctx *Ctx)

	// [Required] Target function under test.
	Target func(ctx *Ctx, in InputType) OutputType

	// [Required] List of test cases.
	Cases []FunctionCase[InputType, OutputType]

	// [Optional] Function to execute before calling the target function. This is called when no BeforeCall function
	// is provided by the the case itself.
	BeforeCall func(ctx *Ctx, in InputType)

	// [Optional] Function to check the output of the target function. This is called when no Check function
	// is provided by the the case itself.
	Check func(ctx *Ctx, in InputType, out OutputType)

	// [Optional] Cleanup function to execute after the test case finishes. This is called when no Cleanup function
	// is provided by the the case itself.
	Cleanup func(ctx *Ctx)

	// [Optional] Teardown function is called after all cases finish
	Teardown func(ctx *Ctx)
}

// Run executes all the test cases in the FunctionMesa instance.
func (m FunctionMesa[I, O]) Run(t *testing.T) {
	im := MethodMesa[any, any, I, O]{
		NewInstance: func(_ *Ctx, _ any) any {
			return nil
		},

		Cases: make([]MethodCase[any, any, I, O], len(m.Cases)),
	}

	checkAndSet(&im.Init, m.Init != nil, func(ctx *Ctx) {
		m.Init(ctx)
	})

	checkAndSet(&im.Target, m.Target != nil, func(ctx *Ctx, _ any, in I) O {
		return m.Target(ctx, in)
	})

	checkAndSet(&im.BeforeCall, m.BeforeCall != nil, func(ctx *Ctx, _ any, in I) {
		m.BeforeCall(ctx, in)
	})

	checkAndSet(&im.Check, m.Check != nil, func(ctx *Ctx, _ any, in I, out O) {
		m.Check(ctx, in, out)
	})

	checkAndSet(&im.Cleanup, m.Cleanup != nil, func(ctx *Ctx, _ any) {
		m.Cleanup(ctx)
	})

	checkAndSet(&im.Teardown, m.Teardown != nil, func(ctx *Ctx) {
		m.Teardown(ctx)
	})

	for i, c := range m.Cases {
		c := c
		im.Cases[i] = MethodCase[any, any, I, O]{
			Name:  c.Name,
			Input: c.Input,
			Skip:  c.Skip,
			InputFn: func(ctx *Ctx, inst any) I {
				return c.InputFn(ctx)
			},
		}

		checkAndSet(&im.Cases[i].BeforeCall, c.BeforeCall != nil, func(ctx *Ctx, _ any, in I) {
			c.BeforeCall(ctx, in)
		})

		checkAndSet(&im.Cases[i].Check, c.Check != nil, func(ctx *Ctx, _ any, in I, out O) {
			c.Check(ctx, in, out)
		})

		checkAndSet(&im.Cases[i].Cleanup, c.Cleanup != nil, func(ctx *Ctx, _ any) {
			c.Cleanup(ctx)
		})
	}

	im.Run(t)
}

// MethodBenchmarkMesa represents a collection of test cases and the functions to create instances
// and execute the target function under test.
type MethodBenchmarkMesa[InstanceType, FieldsType, InputType, OutputType any] struct {
	// [Optional] Function to initialize anything before running the test cases
	Init func(ctx *Ctx)

	// [Required] Function to create a new instance.
	NewInstance func(ctx *Ctx, fields FieldsType) InstanceType

	// [Required] Target function under test.
	Target func(ctx *Ctx, inst InstanceType, in InputType) OutputType

	// [Required] List of test cases.
	Cases []MethodBenchmarkCase[InstanceType, FieldsType, InputType, OutputType]

	// [Optional] Function to check the output of the target function. It will be called instead of the Check function
	// in the MethodMesa if provided.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// [Optional] Function to execute before calling the target function. This is called when no BeforeCall function
	// is provided by the the case itself.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// [Optional] Cleanup function to execute after the test case finishes. This is called when no Cleanup function
	// is provided by the the case itself.
	Cleanup func(ctx *Ctx, inst InstanceType)

	// [Optional] Teardown function is called after all cases finish
	Teardown func(ctx *Ctx)
}

// MethodBenchmarkCase represents a benchmark case with its associated properties.
type MethodBenchmarkCase[InstanceType, FieldsType, InputType, OutputType any] struct {
	// [Required] Name of the benchmark case.
	Name string

	// [Optional] Fields associated with the instance. FieldsFn takes priority over Fields. If fields are not needed
	// to instantiate a the benchmark instance, no fields need to be provided.
	Fields FieldsType

	// [Optional] FieldsFn returns the fields used for this case. FieldsFn takes priority over Fields. If fields are
	// not needed to instantiate a the benchmark instance, no fields need to be provided.
	FieldsFn func(ctx *Ctx) FieldsType

	// [Optional] Input data for the benchmark case. InputFn takes priority over Input. The Input field can be empty if the
	// target function does not take any arguments.
	Input InputType

	// [Optional] InputFn returns the input struct used for this case. It takes priority over the Input field. This can
	// be empty if the target function does not take any arguments.
	InputFn func(ctx *Ctx, inst InstanceType) InputType

	// [Optional] Function to check the output of the target function. It will be called instead of the Check function
	// in the MethodMesa if provided.
	Check func(ctx *Ctx, inst InstanceType, in InputType, out OutputType)

	// [Optional] Reason to skip the benchmark case. The benchmark is only skipped if this field is not empty
	Skip string

	// [Optional] Function to execute before calling the target function. It will be called instead of the BeforeCall
	// function in the MethodMesa if provided.
	BeforeCall func(ctx *Ctx, inst InstanceType, in InputType)

	// [Optional] Cleanup function to execute after the benchmark case finishes. It will be called instead of the Cleanup
	// function in the MethodMesa if provided.
	Cleanup func(ctx *Ctx, inst InstanceType)
}

// Run executes all the benchmark cases in the Mesa instance.
func (m MethodBenchmarkMesa[Inst, F, I, O]) Run(b *testing.B) {
	ctx := newCtx(b)

	if m.Init != nil {
		m.Init(ctx)
	}

	if m.Teardown != nil {
		defer m.Teardown(ctx)
	}

	var result O

	for _, bb := range m.Cases {
		b.Run(bb.Name, func(b *testing.B) {
			if bb.Skip != "" {
				b.Skip(bb.Skip)
			}

			ctx := newCtx(b)

			if bb.FieldsFn != nil {
				bb.Fields = bb.FieldsFn(ctx)
			}

			inst := m.NewInstance(ctx, bb.Fields)

			if bb.InputFn != nil {
				bb.Input = bb.InputFn(ctx, inst)
			}

			cleanup := func() {}

			switch {
			case bb.Cleanup != nil:
				cleanup = func() { bb.Cleanup(ctx, inst) }
			case m.Cleanup != nil:
				cleanup = func() { m.Cleanup(ctx, inst) }
			}

			b.Cleanup(cleanup)

			switch {
			case bb.BeforeCall != nil:
				bb.BeforeCall(ctx, inst, bb.Input)
			case m.BeforeCall != nil:
				m.BeforeCall(ctx, inst, bb.Input)
			}

			var out O

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				innerOut := m.Target(ctx, inst, bb.Input)
				out = innerOut
			}

			result = out

			b.StopTimer()

			for name, value := range ctx.metrics {
				b.ReportMetric(value/float64(b.N), name)
			}

			switch {
			case bb.Check != nil:
				bb.Check(ctx, inst, bb.Input, out)
			case m.Check != nil:
				m.Check(ctx, inst, bb.Input, out)
			}
		})
	}

	var _ = result
}

// WithCases returns a new instance using the provided cases.
func (m MethodBenchmarkMesa[Inst, F, I, O]) WithCases(
	cases []MethodBenchmarkCase[Inst, F, I, O],
) MethodBenchmarkMesa[Inst, F, I, O] {
	m.Cases = cases
	return m
}

func checkAndSet[T any](dst *T, shouldUpdate bool, val T) {
	if shouldUpdate {
		*dst = val
	}
}
