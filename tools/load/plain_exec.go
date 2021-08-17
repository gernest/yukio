package main

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var _ ExecContext = (*Plain)(nil)

type Plain struct {
	source Source
	cancel context.CancelFunc
}

func NewPlain() *Plain {
	return &Plain{
		source: NewPaths(
			Base{},
			"/", "/home",
		),
	}
}

func (p *Plain) Name() string {
	return "Plain"
}

func (p *Plain) Cancel() {
	p.cancel()
}

func (p *Plain) Run(ctx context.Context, log *zap.Logger) {
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	ctx, p.cancel = context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			res, err := target.Do(ctx, p.source.Next())
			if err != nil {
				log.Error("Failed to send event")
				continue
			}
			res.Body.Close()
			if res.StatusCode != http.StatusOK {
				log.Error("Unexpected status", zap.Int("code", res.StatusCode))
			}
		}
	}
}
