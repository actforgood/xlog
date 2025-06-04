// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xlog/blob/main/LICENSE.

package xlog

import "context"

// loggerCtxKey is the context key where logger is stored.
type loggerCtxKey struct{}

// ContextWithLogger returns a new context enriched with logger.
func ContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

// LoggerFromContext returns the logger stored in the context,
// or a [NopLogger] if none is present in the context.
func LoggerFromContext(ctx context.Context) Logger {
	if logger, found := ctx.Value(loggerCtxKey{}).(Logger); found {
		return logger
	}

	return NopLogger{}
}
