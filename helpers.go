package mesa

// MustAssert asserts the type of the given value and fails the test if the input cannot be asserted to the type.
func MustAssert[T any](ctx *Ctx, in any) T {
	val, ok := in.(T)
	ctx.Re.Truef(ok, "Cannot assert type: %v => %v", in, *new(T))
	return val
}
