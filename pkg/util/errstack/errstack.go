package errstack

import (
	"fmt"

	"github.com/pkg/errors"
)

// Println prints err to standard output.
// If err represents a chain of errors, Println prints the first error in the chain that has a stack trace.
func Println(err error) {
	fmt.Println(err)
	firstWithStack := FirstWithStack(err)
	if hasStack(firstWithStack) {
		var st stackTracer
		errors.As(firstWithStack, &st)
		fmt.Printf("First error with a stack trace:\n%v%+v\n", firstWithStack, st.StackTrace())
	}
}

// FirstWithStack returns the first error in the chain of errors that has a stack trace associated with it.
// The chain of errors is determined by repeatedly calling Unwrap() on the err.
// "First" is defined as the error that is returned by the last call to Unwrap().
func FirstWithStack(err error) error {
	if !isWrapper(err) {
		return err
	}

	earlier := FirstWithStack(errors.Unwrap(err))
	if hasStack(earlier) {
		return earlier
	}

	return err
}

// isWrapper returns true if err implements the Unwrap() method.
func isWrapper(err error) bool {
	type wrapper interface {
		Unwrap() error
	}
	_, ok := err.(wrapper) //nolint:errorlint
	return ok
}

// hasStack returns true if err implements the StackTrace() method.
func hasStack(err error) bool {
	_, ok := err.(stackTracer) //nolint:errorlint
	return ok
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
