package werr

import (
	"context"
	"fmt"
)

// WrappedError represents a standard error
// that can also encapsulate a context.
type WrappedError struct {
	Ctx context.Context
	Err error
}

func (w *WrappedError) Error() string {
	return fmt.Sprintf("%s", w.Err)
}

func Wrap(err error, ctx context.Context) *WrappedError {
	return &WrappedError{
		Ctx: ctx,
		Err: err,
	}
}

func (w *WrappedError) WithContext(ctx context.Context) *WrappedError {
	return &WrappedError{
		Ctx: ctx,
		Err: w.Err,
	}
}

func (w *WrappedError) Unwrap() error {
	return w.Err
}
