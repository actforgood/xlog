package xlog_test

import (
	"context"
	"io"
	"testing"

	"github.com/actforgood/xlog"
)

func TestContextWithLogger(t *testing.T) {
	t.Parallel()

	// arrange
	ctx := context.Background()
	expectedLogger := xlog.NewSyncLogger(io.Discard)
	defer expectedLogger.Close()
	expectedDefaultLogger := xlog.NopLogger{}

	// act
	defaultLogger := xlog.LoggerFromContext(ctx)
	newCtx := xlog.ContextWithLogger(ctx, expectedLogger)
	actualLogger := xlog.LoggerFromContext(newCtx)

	// assert
	assertEqual(t, expectedDefaultLogger, defaultLogger)
	assertEqual(t, expectedLogger, actualLogger)
}
