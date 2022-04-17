package domain_service

import (
	"context"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"git.garena.com/xinlong.wu/zoo/util"
)

var UserDomainService = &userDomainService{}

type userDomainService struct {
}

func (u *userDomainService) Register(username, password string) error {
	user := &domain.User{
		Username: username,
	}
	if err := user.CheckUsername(); err != nil {
		return err
	}
	if err := user.CheckPassword(password); err != nil {
		return err
	}
	user.SetPassword(password)
	if err := repository.UserRepository.Create(user); err != nil {
		return err
	}
	return nil
}

func (u *userDomainService) UpdateProfile(username, nickname, avatar string) (*domain.User, error) {
	if nickname == "" || avatar == "" {
		return nil, util.Err.ServerError(err_const.InvalidParam, "username or avatar should not be empty")
	}
	user, err := repository.UserRepository.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	user.Nickname = nickname
	user.Avatar = avatar
	if err := repository.UserRepository.Update(user); err != nil {
		return nil, err
	}
	infra.RedisClient.Del(context.Background(), user.GetProfileRedisKey())

	return user, nil
}
