package main

import (
	"context"
	"flag"
	"sync"
	"time"

	"go.uber.org/zap"
)

var target Endpoint

func main() {
	d := flag.Duration("d", 10*time.Minute, "the length of load testing duration")
	t := flag.String("t", "http://localhost:8090/api/events", "events endpoint")
	flag.Parse()
	target = Endpoint(*t)
	ctx, _ := context.WithTimeout(context.Background(), *d)
	Run(ctx, NewPlain())
}

type ExecContext interface {
	Name() string
	Run(ctx context.Context, log *zap.Logger)
	Cancel()
}

func Run(ctx context.Context, ls ...ExecContext) {
	var wg sync.WaitGroup
	for _, e := range ls {
		wg.Add(1)
		go exec(ctx, e, &wg)
	}
	wg.Wait()
}

func exec(ctx context.Context, e ExecContext, wg *sync.WaitGroup) {
	l := zap.L().Named(e.Name())
	l.Info("Start execution")
	defer func() {
		wg.Done()
		l.Info("Done")
	}()
	safe := make(chan struct{}, 1)
	go func() {
		e.Run(ctx, l)
		safe <- struct{}{}
	}()
	for {
		select {
		case <-safe:
			return
		case <-ctx.Done():
			l.Info("Context was cancelled exiting execution context")
			e.Cancel()
		}
	}
}
