package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/jinzhu/copier"
	"time"
)

type LoginApp struct {
}

func (*LoginApp) Login(req api.LoginReq, resp *api.TokenResp) error {
	user := &domain.User{}
	username := req.Username
	user.Username = username
	if err := user.CheckUsername(); err != nil {
		return err
	}
	if err := user.CheckPassword(req.Password); err != nil {
		return err
	}
	user, err := repository.UserRepository.GetUserByName(username)
	if err != nil {
		return err
	}
	if !util.Encrypt.IsPasswordMatch(user.HashedPassword, req.Password) {
		return util.Err.ServerError(err_const.InvalidPassword, "password is wrong")
	}

	// create token, save token to redis, return token
	token := domain.NewToken(username)
	if err = repository.TokenRepository.Save(token); err != nil {
		return err
	}
	copier.Copy(resp, token)
	return nil
}

func (*LoginApp) Logout(req api.LogoutReq, success *bool) error {
	if err := util.Validator.CheckLength(req.Token, "token", 10, 60); err != nil {
		return util.Err.ServerError(err_const.InvalidParam, err.Error())
	}
	_, err := repository.TokenRepository.Get(req.Token)
	if err != nil {
		return err
	}
	err = repository.TokenRepository.Delete(req.Token)
	if err == nil {
		*success = true
		return nil
	}
	return err
}

func (*LoginApp) RefreshToken(req api.RefreshTokenReq, resp *api.TokenResp) error {
	if err := util.Validator.CheckLength(req.Token, "token", 10, 60); err != nil {
		return util.Err.ServerError(err_const.InvalidParam, err.Error())
	}
	token, err := repository.TokenRepository.Get(req.Token)
	if err != nil {
		return err
	}
	if req.RefreshToken != token.RefreshToken {
		repository.TokenRepository.Delete(token.Token)
		return util.Err.ServerError(err_const.LoginRequired, "invalid refresh token")
	}

	now := time.Now().Unix()
	if now > token.RefreshExpireTime {
		return util.Err.ServerError(err_const.LoginRequired, "expired")
	}

	newToken := domain.NewToken(token.Username)
	if err := repository.TokenRepository.Save(newToken); err != nil {
		return util.Err.ServerError(err_const.LoginRequired, err.Error())
	}

	copier.Copy(resp, newToken)
	return nil
}
