package db

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/models"
)

func TestUsers(t *testing.T) {
	o := badger.DefaultOptions("")
	o.Logger = nil
	o.InMemory = true
	db, err := badger.Open(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	ctx := SetStore(context.Background(), db)
	t.Run("Create", func(t *testing.T) {
		email := "test@yukio.io"
		password := "Pass"
		username := "yukio"
		id, err := CreateUser(ctx, &models.UserData{
			Username: username,
			Email:    email,
			Password: password,
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("GetByID", func(t *testing.T) {
			_, err := GetUserByID(ctx, id)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("GetByEmail", func(t *testing.T) {
			_, err := GetUserByEmail(ctx, email)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("GetAndVerifyByEmail", func(t *testing.T) {
			_, err := GetAndVerifyUserByEmail(ctx, email, password)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("Delete", func(t *testing.T) {
			err := DeleteUser(ctx, id)
			if err != nil {
				t.Error(err)
				return
			}
			// make suer we cant get the user
			_, err = GetUserByID(ctx, id)
			if err == nil {
				t.Error("Expected an error with missing record")
			}
		})
	})
}
