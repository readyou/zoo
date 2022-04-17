package repository

import (
	"github.com/jinzhu/copier"
	"log"
	"time"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
)

var UserRepository *userRepository = &userRepository{}

type userRepository struct {
}

func (*userRepository) Create(user *domain.User) error {
	_, err := infra.XDB.Insert(user)
	//now := time.Now().Unix()
	//_, err := infra.DB.Exec("INSERT INTO `user` (`username`,`hashed_password`,`nickname`,`avatar`,`create_time`,`update_time`) VALUES (?,?,?,?,?,?)",
	//	user.Username, user.HashedPassword, user.Nickname, user.Avatar, now, now)
	//_, err = infra.DB.Exec("select now()")
	if err != nil {
		log.Printf("Insert user error: %+v\n", err)
		return err
	}
	return nil
}

func (*userRepository) Update(user *domain.User) error {
	if _, err := infra.XDB.ID(user.Id).Update(user); err != nil {
		log.Printf("Update user error: %+v\n", err)
		return err
	}
	return nil
}

func (*userRepository) GetByUsername(username string) (user *domain.User, err error) {
	user = &domain.User{Username: username}
	isExist, err := infra.XDB.Get(user)
	if err != nil {
		return nil, err
	}
	if !isExist {
		err = util.Err.ServerError(err_const.UserNotExists, "user not found: "+user.Username)
	}
	return
}

func (*userRepository) GetCache(username string) (user *domain.User, err error) {
	user = &domain.User{
		Username: username,
	}
	key := user.GetProfileRedisKey()
	err = infra.RedisUtil.GetOrSet(key, func(key string, ret any) error {
		user, err = UserRepository.GetByUsername(username)
		if err != nil {
			return err
		}
		copier.Copy(ret, user)
		return nil
	}, time.Hour, user)
	return
}
