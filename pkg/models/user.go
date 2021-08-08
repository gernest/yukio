package models

import (
	valid "github.com/asaskevich/govalidator"
)

type UserData struct {
	Username string `valid:"alphanum,required"`
	Email    string `valid:"email,required"`
	Password string `valid:"alphanum,required"`
}

func (u UserData) Valid() error {
	_, err := valid.ValidateStruct(u)
	return err
}

type User struct {
	ID int64

	Username string
	Email    string
}
