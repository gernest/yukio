package loga

import (
	"context"

	"go.uber.org/zap"
)

type logaKey struct{}

func Get(ctx context.Context) *zap.Logger {
	if k := ctx.Value(logaKey{}); k != nil {
		return k.(*zap.Logger)
	}
	return zap.L()
}

func Set(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logaKey{}, l)
}
