package tt

import "runtime/debug"

// PanicHandler will run the input function and recover
// from any panics it might generate.
//
// It will then save the panic payload and return it
// so it can be asserted by other functions on the test.
func PanicHandler(fn func()) (panicPayload interface{}, stacktrace string) {
	defer func() {
		// Overwrites the panic payload if a pannic actually occurs:
		if r := recover(); r != nil {
			panicPayload = r
			stacktrace = string(debug.Stack())
		}
	}()

	fn()
	return nil, ""
}
