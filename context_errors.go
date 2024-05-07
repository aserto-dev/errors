package errors

import (
	"context"
)

// ContextError represents a standard error
// that can also encapsulate a context.
type ContextError struct {
	Err error
	Ctx context.Context
}

func WrapWithContext(err error, ctx context.Context) *ContextError {
	return &ContextError{
		Err: err,
		Ctx: ctx,
	}
}

func (ce *ContextError) Error() string {
	return ce.Err.Error()
}

func (ce *ContextError) Cause() error {
	return ce.Err
}

func (ce *ContextError) Unwrap() error {
	return ce.Err
}
