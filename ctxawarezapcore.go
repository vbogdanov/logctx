package logctx

import (
	"context"

	"go.uber.org/zap/zapcore"
)

const fieldName = "context"

// CtxField wraps context.Context as a Field. To be used by CtxAwareZapCore.
func CtxField(ctx context.Context) zapcore.Field {
	return zapcore.Field{Key: fieldName, Type: zapcore.ReflectType, Interface: ctx}
}

// OnLogWrite wraps a function defining behavior when a log is written if context.Context is added as field.
// The intention is to integrate tracing and logging. For example
// * opencensus span context can be added as fields
// * record of a log message can be added in the trace
// It is OK to mutate and return the passed fields slice.
// This function is invoked after Level check and only if the log entry is about to be written out.
type OnLogWrite func(ctx context.Context, ent zapcore.Entry, fields []zapcore.Field) []zapcore.Field

// WrapCore allows an easy integration with zap.WrapCore().
// It creates a new CtxAwareZapCore around the passed core (usually zapcore.ioCore)
// and sets the current OnLogWrite in it.
func (fn OnLogWrite) WrapCore(core zapcore.Core) zapcore.Core {
	return &CtxAwareZapCore{
		Core:       core,
		OnLogWrite: fn,
	}
}

// CtxAwareZapCore is a wrapping zapcore.Core implementation that checks for wrapped context.Context
type CtxAwareZapCore struct {
	zapcore.Core
	// latestCtx keeps the latest context participating in the logging
	latestCtx  context.Context
	OnLogWrite OnLogWrite
}

type leveler interface {
	Level() zapcore.Level
}

var (
	_ zapcore.Core         = (*CtxAwareZapCore)(nil)
	_ zapcore.LevelEnabler = (*CtxAwareZapCore)(nil)
	_ leveler              = (*CtxAwareZapCore)(nil)
)

func (s *CtxAwareZapCore) Level() zapcore.Level {
	return zapcore.LevelOf(s.Core)
}

func (s *CtxAwareZapCore) ctxFromSelfOrFields(fields []zapcore.Field) (context.Context, []zapcore.Field) {
	storedFields := make([]zapcore.Field, 0, len(fields))
	ctx := s.latestCtx
	for _, f := range fields {
		if f.Key == fieldName && f.Type == zapcore.ReflectType {
			maybeCtx, ok := f.Interface.(context.Context)
			if ok {
				ctx = maybeCtx
				continue
			}
		}
		storedFields = append(storedFields, f)
	}
	return ctx, storedFields
}

// With is removes context.Context if present in the fields and stores it in the logger core
func (s *CtxAwareZapCore) With(fields []zapcore.Field) zapcore.Core {
	ctx, storedFields := s.ctxFromSelfOrFields(fields)
	return &CtxAwareZapCore{
		Core:       s.Core.With(storedFields),
		OnLogWrite: s.OnLogWrite,
		latestCtx:  ctx,
	}
}

// Check checks if the entry should be logged. The method checks with the wrapped core, but add itself as the writer
func (s *CtxAwareZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !s.Enabled(ent.Level) {
		return ce
	}
	// check downstream is interested
	res := s.Core.Check(ent, nil)
	if res == nil {
		return ce
	}
	// add only self, so we can modify what is written before pushing it downwards
	return ce.AddCore(ent, s)
}

const extraCapacity = 10

// Write writes a given log message. This method calls OnLogWrite callback before
// delegating to the wrapped core.
func (s *CtxAwareZapCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	if s.OnLogWrite == nil {
		return s.Core.Write(ent, fields)
	}
	ctx, fields := s.ctxFromSelfOrFields(fields)
	if ctx == nil {
		return s.Core.Write(ent, fields)
	}
	extendedFields := make([]zapcore.Field, 0, len(fields)+extraCapacity)
	// copy to avoid modifying the original in the callback
	_ = copy(extendedFields, fields)
	extendedFields = s.OnLogWrite(s.latestCtx, ent, extendedFields)
	return s.Core.Write(ent, extendedFields)
}
