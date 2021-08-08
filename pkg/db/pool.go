package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var pool *pgxpool.Pool

func init() {
	var err error
	pool, err = pgxpool.Connect(context.Background(), os.Getenv("POSTGRESQL_URL"))
	if err != nil {
		panic("Can't connect to timescale db" + err.Error())
	}
}

// Do acquires a new connection and calls fn with it. Ensures connection is
// released before exit.
func Do(ctx context.Context, fn func(ctx context.Context, conn *pgxpool.Conn) error) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return fn(ctx, conn)
}
