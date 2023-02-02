package logctx

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type enhancedError struct {
	wrapped error
	logger  *zap.Logger
}

// NewError creates a new `error` using the provided message and wraps in logging context enhanced error.
func NewError(ctx context.Context, msg string, fields ...zap.Field) error {
	return &enhancedError{
		wrapped: errors.New(msg),
		logger:  From(ctx).With(fields...),
	}
}

// EnhanceError wraps the provided error in logging context enhanced error.
// Optionally, fields are added to the logging context.
func EnhanceError(ctx context.Context, err error, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	return &enhancedError{
		wrapped: err,
		logger:  From(ctx).With(fields...),
	}
}

// Unwrap is part of go1.13 error extension. Allows the use of errors.Is and errors.As.
func (e *enhancedError) Unwrap() error {
	return e.wrapped
}

// Error method is required for logctx.enhancedError to be an error.
func (e *enhancedError) Error() string {
	return e.wrapped.Error()
}
