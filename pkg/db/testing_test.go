package db

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v3"
)

func Setup(t *testing.T) context.Context {
	t.Helper()
	o := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(nil)
	store, err := badger.Open(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		store.Close()
	})
	return SetStore(context.Background(), store)
}
