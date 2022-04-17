package domain_service

import (
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"git.garena.com/xinlong.wu/zoo/util"
)

var TokenDomainService *tokenDomainService = &tokenDomainService{}

type tokenDomainService struct {
}

func (*tokenDomainService) ValidateToken(token string) (*domain.Token, error) {
	if err := util.Validator.CheckLength(token, "token", 10, 60); err != nil {
		return nil, util.Err.ServerError(err_const.InvalidParam, err.Error())
	}
	tok, err := repository.TokenRepository.Get(token)
	if err != nil {
		return nil, err
	}
	if tok.IsRefreshTokenExpired() {
		return nil, util.Err.ServerError(err_const.LoginRequired, "token expired")
	}
	if tok.IsTokenExpired() {
		return nil, util.Err.ServerError(err_const.LoginRequired, "token expired")
	}
	return tok, nil
}

func (*tokenDomainService) Login(username, password string) (*domain.Token, error) {
	user := &domain.User{}
	user.Username = username
	if err := user.CheckUsername(); err != nil {
		return nil, err
	}
	if err := user.CheckPassword(password); err != nil {
		return nil, err
	}
	user, err := repository.UserRepository.GetCache(username)
	if err != nil {
		return nil, err
	}
	if !user.IsPasswordMatch(password) {
		return nil, util.Err.ServerError(err_const.InvalidPassword, "password is wrong")
	}

	// create token, save token to redis, return token
	token := domain.NewToken(username)
	if err = repository.TokenRepository.Save(token); err != nil {
		return nil, err
	}
	return token, nil
}

func (*tokenDomainService) Logout(token string) error {
	_, err := repository.TokenRepository.Get(token)
	if err != nil {
		return err
	}
	return repository.TokenRepository.Delete(token)
}

func (*tokenDomainService) RefreshToken(token, refreshToken string) (*domain.Token, error) {
	if err := util.Validator.CheckLength(token, "token", 10, 60); err != nil {
		return nil, util.Err.ServerError(err_const.InvalidParam, err.Error())
	}
	tok, err := repository.TokenRepository.Get(token)
	if err != nil {
		return nil, err
	}
	if tok.RefreshToken != refreshToken {
		repository.TokenRepository.Delete(token)
		return nil, util.Err.ServerError(err_const.LoginRequired, "invalid refresh token")
	}

	if tok.IsRefreshTokenExpired() {
		return nil, util.Err.ServerError(err_const.LoginRequired, "expired")
	}

	repository.TokenRepository.Delete(token)
	newToken := domain.NewToken(tok.Username)
	if err := repository.TokenRepository.Save(newToken); err != nil {
		return nil, util.Err.ServerError(err_const.LoginRequired, err.Error())
	}

	return newToken, nil
}
