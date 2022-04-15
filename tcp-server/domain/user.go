package domain

import (
	"git.garena.com/xinlong.wu/zoo/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
	"unicode"
)

type User struct {
	Id             int64
	Username       string // user's identifier(unique) name, same as Id
	HashedPassword string
	Nickname       string
	Avatar         string
	CreateTime     int64 `xorm:"created"`
	UpdateTime     int64 `xorm:"updated"`
}

func (*User) UpdateProfile() error {
	return nil
}

func (user *User) CheckUsername() error {
	if err := util.Validator.CheckLength(user.Username, "username", 1, 64); err != nil {
		return err
	}
	if !user.isValidName(user.Username) {
		return util.Err.ServerError(err_const.InvalidParam, "username should be made up of letters, digits, and underscores")
	}
	return nil
}

func (user *User) CheckPassword(password string) error {
	if err := util.Validator.CheckLength(password, "password", 8, 32); err != nil {
		return util.Err.ServerError(err_const.InvalidParam, err.Error())
	}
	return nil
}

func (user *User) EncryptPassword(password string) string {
	return util.Encrypt.EncryptPassword(password)
}

// name should be made up of letters, digits, and underscores
func (*User) isValidName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}
	return true
}