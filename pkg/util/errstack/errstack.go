package errstack

import (
	"errors"
	"fmt"

	es "github.com/go-errors/errors"
)

// ToString returns a string representation of the error err.
// If err represents a chain of errors,
// ToString appends a representation of the first (deepest) error in the chain that has a stack trace, if any.
// The chain of errors is formed by recursively calling Unwrap() on err.
func ToString(err error) string {
	var errWithStack *es.Error
	if errors.As(FirstWithStack(err), &errWithStack) {
		// If there is a stack trace attached to any of the error chain.

		if errWithStack == err { //nolint:errorlint
			// If the first error with a stack trace is the error itself, just print it, including the stack trace.
			return errWithStack.ErrorStack()
		}

		// If the error itself does not have a stack trace attached, but another one in the chain does,
		// first print the error itself and, in addition, print the first (deepest) error with a stack trace.
		return fmt.Sprintf("%v\nFirst error with a stack trace:\n%s", err, errWithStack.ErrorStack())
	}

	// If there is no stack trace anywhere in the error chain, simply print the error.
	return fmt.Sprintf("%v", err)
}

// Println is a convenience function (sugar) that prints the output of ToString(err) to standard output.
func Println(err error) {
	fmt.Println(ToString(err))
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
	_, ok := err.(*es.Error) //nolint:errorlint
	return ok
}
