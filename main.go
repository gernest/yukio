package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/config"
	"github.com/gernest/yukio/pkg/db"
	"github.com/gernest/yukio/pkg/handlers"
	"github.com/gernest/yukio/pkg/loga"
	"github.com/gernest/yukio/pkg/web"
	"github.com/gorilla/mux"
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
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config,c",
			Usage: "Path to configuration file",
		},
	}
	app.Action = run
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(cliCtx *cli.Context) error {
	ctx := context.Background()
	o := config.Default()
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

	ctx = db.SetStore(ctx, store)
	ctx = loga.Set(ctx, zl)
	m := mux.NewRouter()
	m.Use(ContextMiddleware(ctx))
	web.AddRoutes(m)
	handlers.AddRoutes(m)
	zl.Info("Starting server", zap.Int("port", o.ListenPort))
	return http.ListenAndServe(fmt.Sprintf(":%d", o.ListenPort), m)
}

func ContextMiddleware(ctx context.Context) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
