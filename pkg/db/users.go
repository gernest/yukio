package db

import (
	"context"
	"errors"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func UserKeys(u *models.User) (keys []*Key) {
	keys = []*Key{
		gk().UserID(u.Id), gk().UserEmail(u.Email),
	}
	for _, v := range u.Domains {
		keys = append(keys, gk().Domain(v.Name))
	}
	return
}

func CreateUser(ctx context.Context, usr *models.UserData) (id uuid.UUID, err error) {
	var password, value []byte
	password, err = bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	id = models.NewID()
	u := &models.User{
		Id:       id[:],
		Username: usr.Username,
		Email:    usr.Email,
		Password: password,
	}
	value, err = proto.Marshal(u)
	if err != nil {
		return
	}
	err = Set(ctx, value, UserKeys(u)...)
	return
}

func UpdateUser(ctx context.Context, u *models.User) error {
	value, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	return Set(ctx, value, UserKeys(u)...)
}

func DeleteUser(ctx context.Context, id uuid.UUID) error {
	return GetStore(ctx).Update(func(txn *badger.Txn) error {
		k := gk().UserID(id[:])
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
	var u models.User
	err := get(ctx, func(k *Key) { k.UserID(id[:]) }, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := get(ctx, func(k *Key) { k.UserEmail(email) }, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByDomainl(ctx context.Context, domain string) (*models.User, error) {
	var u models.User
	err := get(ctx, func(k *Key) { k.Domain(domain) }, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, badger.ErrKeyNotFound)
}

func get(ctx context.Context, key func(*Key), o protoreflect.ProtoMessage) error {
	k := gk()
	defer pk(k)
	key(k)
	return GetStore(ctx).View(func(txn *badger.Txn) error {
		i, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error { return proto.Unmarshal(val, o) })
	})
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
