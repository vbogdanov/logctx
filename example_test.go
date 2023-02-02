package logctx_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vbogdanov/logctx"
	"go.uber.org/zap"
)

func ExampleForError() {
	logctx.DefaultLogger = zap.NewExample()
	ctx := context.TODO()

	err := DoOperation(ctx)
	if err != nil {
		logctx.ForError(ctx, err).Error("operation failed")
	}

	// Output:
	// {"level":"debug","msg":"starting operation","username":"random"}
	// {"level":"debug","msg":"starting in depth","username":"random","depth":2}
	// {"level":"error","msg":"operation failed","username":"random","depth":2,"of test":"1s"}
}

func DoOperation(ctx context.Context) error {
	ctx = logctx.WithFields(ctx,
		zap.String("username", "random"),
	)
	logctx.From(ctx).Debug("starting operation")
	// ...
	err := DoInDepth(ctx)
	if err != nil {
		return fmt.Errorf("wrapped with errorf: %w", err)
	}
	return nil
}

func DoInDepth(ctx context.Context) error {
	ctx = logctx.WithFields(ctx,
		zap.Int("depth", 2),
	)
	logctx.From(ctx).Debug("starting in depth")
	// ...
	err := failingOp()
	if err != nil {
		// add the most possible context available in the error
		return logctx.EnhanceError(ctx, err, zap.Duration("of test", 1*time.Second))
	}
	return nil
}

func failingOp() error {
	return errors.New("something")
}
