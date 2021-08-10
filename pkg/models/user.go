package models

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/google/uuid"
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

func NewID() uuid.UUID {
	return uuid.New()
}
