package config

import (
	"context"
	"time"

	"github.com/dgraph-io/badger/v3"
)

const DefaultSessionWindow = 3 * time.Minute
const DefaultFlushInterval = time.Second
const DefaultListenPort = 8090

type Config struct {
	SessionWindow time.Duration
	TimeSeries    TimeSeries
	Store         badger.Options
	ListenPort    int
}

func Default() Config {
	return Config{
		SessionWindow: DefaultSessionWindow,
		TimeSeries: TimeSeries{
			FlushInterval: DefaultFlushInterval,
		},
		Store:      badger.DefaultOptions("./data"),
		ListenPort: DefaultListenPort,
	}
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
