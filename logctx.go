// Package logctx, short for Logging Context, provides utils to keep additional logging context in context.Context
// and use it.
package logctx

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type keyType struct{}

var (
	key = &keyType{}

	// exporter global settings. Modify only during initialization:

	// DefaultLogger is the logger to use if the provided ctx does not contain one.
	DefaultLogger = zap.NewNop()
	// AddCtxFields defines whether the context to be added as a Field in the log to be returned.
	AddCtxFields = false
)

func newCtx(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, key, l)
}

// From provides a *zap.Logger from the given context.
// New logger is created if one is not associated with the context.
func From(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(key).(*zap.Logger)
	if !ok {
		return DefaultLogger
	}
	if AddCtxFields {
		l = l.With(CtxField(ctx))
	}
	return l
}

// Debug simplifies logging simple debug messages.
// NB: if ctx == nil, DefaultLogger is used instead of From(ctx).
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx == nil {
		DefaultLogger.Debug(msg, fields...)
	}
	From(ctx).Debug(msg, fields...)
}

// Sugar provides a *zap.SugaredLogger from the given context.
// New logger is created if one is not associated with the context.
func Sugar(ctx context.Context) *zap.SugaredLogger {
	return From(ctx).Sugar()
}

// With enhances the logging context with the given args. Similar to the [*zap.SugaredLogger] With method.
func With(ctx context.Context, args ...interface{}) context.Context {
	return newCtx(ctx, Sugar(ctx).With(args...).Desugar())
}

// WithFields enhances the logging context with the provided fields.
// See [go.uber.org/zap.Logger#With] method
//
// [go.uber.org/zap.Logger#With]: https://pkg.go.dev/go.uber.org/zap@v1.24.0#Logger.With
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	return newCtx(ctx, From(ctx).With(fields...))
}

// ForError provides a new [*zap.Logger] with the error already added as Field.
// See [From] method.
func ForError(ctx context.Context, err error) *zap.Logger {
	var (
		l    *zap.Logger
		elog *enhancedError
	)
	if errors.As(err, &elog) {
		l = elog.logger
		if AddCtxFields {
			l = l.With(CtxField(ctx))
		}
	} else {
		l = From(ctx)
	}
	l.With(zap.Error(err))
	return l
}
