package db

import (
	"context"
	"crypto/rand"

	"github.com/dchest/siphash"
	"github.com/dgraph-io/badger/v3"
)

func getRandom(ctx context.Context, key *Key) error {
	k := gk()
	defer pk(k)
	k.Hash()
	return GetStore(ctx).Update(func(txn *badger.Txn) error {
		i, err := txn.Get(k.Bytes())
		if err != nil {
			if IsNotFound(err) {
				r := make([]byte, 16)
				_, err := rand.Read(r)
				if err != nil {
					return err
				}
				key.Write(r)
				return txn.Set(k.Bytes(), r)
			}
			return err
		}
		return i.Value(func(val []byte) error {
			key.Write(val)
			return nil
		})
	})
}

func GenerateUserID(ctx context.Context, remoteIP, userAgent, domain, hostname string) (id uint64, err error) {
	k := gk()
	defer pk(k)
	err = getRandom(ctx, k)
	if err != nil {
		return
	}
	ws := gk()
	defer pk(ws)
	sh := siphash.New(k.Bytes())
	_, err = sh.Write(ws.Bytes())
	if err != nil {
		return
	}
	id = sh.Sum64()
	return
}
