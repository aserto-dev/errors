package errors

import (
	"context"
	"fmt"
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
	return fmt.Sprintf("%s", ce.Err)
}

func (ce *ContextError) Unwrap() error {
	return ce.Err
}
