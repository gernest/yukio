package db

import (
	"bytes"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

var pool = &sync.Pool{
	New: func() interface{} { return new(Key) },
}

var (
	UsersID    = []byte("/users/i/")
	UsersEmail = []byte("/users/e/")
)

type Key struct {
	bytes.Buffer
}

func (k *Key) UserID(id uuid.UUID) *Key {
	k.Write(UsersID)
	k.Write(id[:])
	return k
}

func (k *Key) UserEmail(email string) *Key {
	k.Write(UsersEmail)
	k.WriteString(email)
	return k
}

func gk() *Key {
	return pool.Get().(*Key)
}

func pk(k *Key) {
	k.Reset()
	pool.Put(k)
}

var db *badger.DB

func init() {
	opts := badger.DefaultOptions(os.Getenv("STORAGE_PATH"))
	var err error
	db, err = badger.Open(opts)
	if err != nil {
		panic("Failed to open database")
	}
}

func Set(value []byte, keys ...*Key) error {
	defer func() {
		for _, k := range keys {
			pk(k)
		}
	}()
	return db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Set(key.Bytes(), value); err != nil {
				return nil
			}
		}
		return nil
	})
}
