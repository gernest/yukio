package config

import (
	"context"
	"time"
)

const DefaultSessionWindow = 3 * time.Minute

type Config struct {
	SessionWindow time.Duration
	TimeSeries    TimeSeries
}

type TimeSeries struct {
	FlushInterval time.Duration
}

type configKey struct{}

func Set(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, configKey{}, c)
}

func Get(ctx context.Context) Config {
	return ctx.Value(configKey{}).(Config)
}
