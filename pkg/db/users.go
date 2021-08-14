package db

import (
	"context"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func CreateUser(ctx context.Context, usr *models.UserData) (id uuid.UUID, err error) {
	var password, value []byte
	password, err = bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	id = models.NewID()
	value, err = proto.Marshal(&models.User{
		Id:       id[:],
		Username: usr.Username,
		Email:    usr.Email,
		Password: password,
	})
	if err != nil {
		return
	}
	err = Set(ctx, value, gk().UserID(id), gk().UserEmail(usr.Email))
	return
}

func DeleteUser(ctx context.Context, id uuid.UUID) error {
	return GetStore(ctx).Update(func(txn *badger.Txn) error {
		k := gk().UserID(id)
		defer pk(k)
		i, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		var u models.User
		err = i.Value(func(val []byte) error {
			return proto.Unmarshal(val, &u)
		})
		if err != nil {
			return err
		}
		ek := gk().UserEmail(u.Email)
		defer pk(ek)
		if err := txn.Delete(k.Bytes()); err != nil {
			return err
		}
		return txn.Delete(ek.Bytes())
	})
}

func GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	k := gk().UserID(id)
	defer pk(k)
	var u models.User
	err := GetStore(ctx).View(func(txn *badger.Txn) error {
		i, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error { return proto.Unmarshal(val, &u) })
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	k := gk().UserEmail(email)
	defer pk(k)
	var u models.User
	err := GetStore(ctx).View(func(txn *badger.Txn) error {
		i, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error { return proto.Unmarshal(val, &u) })
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetAndVerifyUserByEmail(ctx context.Context, email, password string) (*models.User, error) {
	u, err := GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return nil, err
	}
	return u, nil
}
