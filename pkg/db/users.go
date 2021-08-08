package db

import (
	"context"
	"encoding/base64"

	"github.com/gernest/yukio/pkg/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(ctx context.Context, usr *models.UserData) (id int64, err error) {
	o, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	pass := base64.StdEncoding.EncodeToString(o)
	err = Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		row := conn.QueryRow(ctx, `insert into users (username,email,password) values($1,$2,$3) RETURNING id;`,
			usr.Username, usr.Email, pass)
		return row.Scan(&id)
	})
	return
}

func DeleteUser(ctx context.Context, id int64) error {
	return Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		_, err := conn.Exec(ctx, `delete from users where id =$1;`, id)
		return err
	})
}

func GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var usr models.User
	err := Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		return conn.QueryRow(ctx, `select username,email from users where id=$1;`, id).
			Scan(
				&usr.Username,
				&usr.Email,
			)
	})
	if err != nil {
		return nil, err
	}
	usr.ID = id
	return &usr, nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var usr models.User
	err := Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		return conn.QueryRow(ctx, `select id,username from users where email=$1;`, email).
			Scan(
				&usr.ID,
				&usr.Username,
			)
	})
	if err != nil {
		return nil, err
	}
	usr.Email = email
	return &usr, nil
}

func GetAndVerifyUserByEmail(ctx context.Context, email, password string) (*models.User, error) {
	var usr models.User
	var passwd string
	err := Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		return conn.QueryRow(ctx, `select id,username,password from users where email=$1;`, email).
			Scan(
				&usr.ID,
				&usr.Username,
				&passwd,
			)
	})
	if err != nil {
		return nil, err
	}
	storedHash, err := base64.StdEncoding.DecodeString(passwd)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(storedHash, []byte(password)); err != nil {
		return nil, err
	}
	usr.Email = email
	return &usr, nil
}
