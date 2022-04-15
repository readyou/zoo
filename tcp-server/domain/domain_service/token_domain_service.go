package domain_service

import (
	"git.garena.com/xinlong.wu/zoo/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"git.garena.com/xinlong.wu/zoo/util"
)

var TokenDomainService *tokenDomainService = &tokenDomainService{}

type tokenDomainService struct {
}

func (*tokenDomainService) ValidateToken(token string) (*domain.Token, error) {
	tok, err := repository.TokenRepository.Get(token)
	if err != nil {
		return nil, err
	}
	if tok.IsRefreshTokenExpired() {
		return nil, util.Err.ServerError(err_const.LoginRequired, "token expired")
	}
	if tok.IsTokenExpired() {
		return nil, util.Err.ServerError(err_const.TokenExpired, "token expired")
	}
	return tok, nil
}
