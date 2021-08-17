package db

import (
	"bytes"
	"context"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

var pool = &sync.Pool{
	New: func() interface{} { return new(Key) },
}

var (
	UsersID     = []byte("/u/i/")
	UsersEmail  = []byte("/u/e/")
	SiteSession = []byte("/s/s/")
	Domains     = []byte("/d/")
)

type Key struct {
	bytes.Buffer
}

func (k *Key) UserID(id []byte) *Key {
	k.Write(UsersID)
	k.Write(id[:])
	return k
}

func (k *Key) UserEmail(email string) *Key {
	k.Write(UsersEmail)
	k.WriteString(email)
	return k
}

func (k *Key) SessionID(userID []byte, domain string) *Key {
	k.Write(SiteSession[:])
	k.Write(userID[:])
	k.WriteString(domain)
	return k
}

func (k *Key) Domain(domain string) *Key {
	k.Write(Domains)
	k.WriteString(domain)
	return k
}

func gk() *Key {
	return pool.Get().(*Key)
}

func pk(k *Key) {
	k.Reset()
	pool.Put(k)
}

type dbKey struct{}

func GetStore(ctx context.Context) *badger.DB {
	return ctx.Value(dbKey{}).(*badger.DB)
}

func SetStore(ctx context.Context, s *badger.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, s)
}

func Set(ctx context.Context, value []byte, keys ...*Key) error {
	defer func() {
		for _, k := range keys {
			pk(k)
		}
	}()
	return GetStore(ctx).Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Set(key.Bytes(), value); err != nil {
				return nil
			}
		}
		return nil
	})
}
