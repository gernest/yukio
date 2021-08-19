package config

import (
	"context"
	"net/url"
	"time"

	"github.com/dgraph-io/badger/v3"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/urfave/cli"
)

const DefaultSessionWindow = 3 * time.Minute
const DefaultFlushInterval = time.Second
const DefaultListenPort = 8090

type Config struct {
	SessionWindow time.Duration
	TimeSeries    TimeSeries
	Store         badger.Options
	Remote        Remote
	ListenPort    int
}

type QueryConfig struct {
	Limits QueryLimits
}

type QueryLimits struct {
	Sample       int
	Concurrency  int
	BytesInFrame int
}

type Remote struct {
	Read  RemoteConfig
	Write RemoteConfig
	Query QueryConfig
}

type RemoteConfig struct {
	URL              string
	Timeout          time.Duration
	Headers          map[string]string
	RetryOnRateLimit bool
}

func (r *RemoteConfig) Config() (*remote.ClientConfig, error) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return nil, err
	}
	return &remote.ClientConfig{
		URL:              &config_util.URL{URL: u},
		Timeout:          model.Duration(r.Timeout),
		Headers:          r.Headers,
		RetryOnRateLimit: r.RetryOnRateLimit,
	}, nil
}

func (c Config) With(ctx *cli.Context) Config {
	c.Remote.Read.URL = ctx.GlobalString("remote-read-url")
	c.Remote.Read.Timeout = ctx.GlobalDuration("remote-read-timeout")
	c.Remote.Read.RetryOnRateLimit = ctx.GlobalBool("remote-read-retry-onlimit")

	c.Remote.Write.URL = ctx.GlobalString("remote-write-url")
	c.Remote.Write.Timeout = ctx.GlobalDuration("remote-write-timeout")
	c.Remote.Write.RetryOnRateLimit = ctx.GlobalBool("remote-write-retry-onlimit")
	return c
}

func Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "config,c",
			Usage:  "Path to configuration file",
			EnvVar: "Y_CONFIG",
		},

		cli.StringFlag{
			Name:   "remote-read-url",
			Usage:  "remote storate url for reads",
			EnvVar: "Y_REMOTE_READ_URL",
		},
		cli.StringFlag{
			Name:   "remote-read-timeout",
			Usage:  "timeouts for remote  reads",
			EnvVar: "Y_REMOTE_READ_TIMEOUT",
		},
		cli.BoolFlag{
			Name:   "remote-read-retry-onlimi",
			Usage:  "retry on rate limit on remote reads",
			EnvVar: "Y_REMOTE_READ_RETRY_ON_LIMIT",
		},
		cli.StringFlag{
			Name:   "remote-write-url",
			Usage:  "remote storate url for writes",
			EnvVar: "Y_REMOTE_WRITE_URL",
		},
		cli.DurationFlag{
			Name:   "remote-write-timeout",
			Usage:  "timeouts for remote  writes",
			EnvVar: "Y_REMOTE_WRITE_TIMEOUT",
			Value:  time.Minute,
		},
		cli.BoolFlag{
			Name:   "remote-write-retry-onlimi",
			Usage:  "retry on rate limit on remote writes",
			EnvVar: "Y_REMOTE_WRITE_RETRY_ON_LIMIT",
		},
	}
}

func Default() Config {
	return Config{
		SessionWindow: DefaultSessionWindow,
		TimeSeries: TimeSeries{
			FlushInterval: DefaultFlushInterval,
		},
		Store:      badger.DefaultOptions("/data"),
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
