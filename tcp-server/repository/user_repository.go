package repository

import (
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
	"log"
)

var UserRepository *userRepository = &userRepository{}

type userRepository struct {
}

func (*userRepository) Create(user *domain.User) error {
	if _, err := infra.DB.Insert(user); err != nil {
		log.Printf("Insert user error: %+v\n", err)
		return err
	}
	return nil
}

func (*userRepository) Update(user *domain.User) error {
	if _, err := infra.DB.ID(user.Id).Update(user); err != nil {
		log.Printf("Update user error: %+v\n", err)
		return err
	}
	return nil
}

func (*userRepository) GetUserByName(username string) (user *domain.User, err error) {
	user = &domain.User{Username: username}
	isExist, err := infra.DB.Get(user)
	if err != nil {
		return nil, err
	}
	if !isExist {
		err = util.Err.ServerError(err_const.UserNotExists, "user not found")
	}
	return
}
