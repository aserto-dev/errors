package errors

import (
	"context"

	"github.com/pkg/errors"
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

func WrapConext(err error, ctx context.Context, message string) *ContextError {
	return WrapWithContext(errors.Wrap(err, message), ctx)
}

func WrapfConext(err error, ctx context.Context, format string, args ...interface{}) *ContextError {
	return WrapWithContext(errors.Wrapf(err, format, args...), ctx)
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
