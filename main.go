package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/config"
	"github.com/gernest/yukio/pkg/db"
	"github.com/gernest/yukio/pkg/events"
	"github.com/gernest/yukio/pkg/handlers"
	"github.com/gernest/yukio/pkg/loga"
	"github.com/gernest/yukio/pkg/web"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

//go:generate go run tools/tracker/main.go
//go:generate go run tools/refdb/main.go
//go:generate protoc  --go_out=. --go_opt=paths=source_relative   pkg/models/models.proto
func main() {
	app := cli.NewApp()
	app.Name = "yukio"
	app.Usage = "Web analytics"
	app.Flags = config.Flags()
	app.Action = run
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(cliCtx *cli.Context) error {
	ctx := context.Background()
	o := config.Default().With(cliCtx)
	if f := cliCtx.GlobalString("config"); f != "" {
		n, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("Failed to open config file path:%s err:%v ", f, err)
		}
		err = json.NewDecoder(n).Decode(&o)
		n.Close()
		if err != nil {
			return fmt.Errorf("Failed to decode config file path:%s err:%v ", f, err)
		}
	}

	store, err := badger.Open(o.Store)
	if err != nil {
		return fmt.Errorf("Failed to open  storage  err:%v ", err)
	}
	defer store.Close()
	zc := zap.NewProductionConfig()
	zc.DisableStacktrace = true
	zl, err := zc.Build(zap.WithCaller(false))
	if err != nil {
		return err
	}
	defer zl.Sync()

	// set values passed through the context
	ctx = db.SetStore(ctx, store)
	ctx = loga.Set(ctx, zl)
	ctx = config.Set(ctx, o)

	writeConfig, err := o.Remote.Write.Config()
	if err != nil {
		return fmt.Errorf("Failed to decode  remote store write config  err:%v ", err)
	}
	write, err := remote.NewWriteClient("yukio-promscale", writeConfig)
	if err != nil {
		return fmt.Errorf("Failed to create  remote store write client  err:%v ", err)
	}

	readConfig, err := o.Remote.Read.Config()
	if err != nil {
		return fmt.Errorf("Failed to decode  remote store read config  err:%v ", err)
	}
	read, err := remote.NewReadClient("yukio-promscale", readConfig)
	if err != nil {
		return fmt.Errorf("Failed to create  remote store write client  err:%v ", err)
	}
	ctx = events.SetupQuery(ctx, zl, read)

	m := mux.NewRouter()
	m.Use(ContextMiddleware(ctx))
	web.AddRoutes(m)
	handlers.AddRoutes(m, zl)
	zl.Info("Starting server", zap.Int("port", o.ListenPort))
	go func() {
		events.WriteLoop(ctx, write, o.TimeSeries.FlushInterval)
	}()
	svr := &http.Server{
		Handler:     m,
		Addr:        fmt.Sprintf(":%d", o.ListenPort),
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	return svr.ListenAndServe()
}

func ContextMiddleware(ctx context.Context) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
