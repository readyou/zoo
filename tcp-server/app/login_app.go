package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain/domain-service"
	"github.com/jinzhu/copier"
)

type LoginApp struct {
}

func (*LoginApp) Login(req api.LoginReq, resp *api.TokenResp) error {
	token, err := domain_service.TokenDomainService.Login(req.Username, req.Password)
	if err != nil {
		return err
	}
	copier.Copy(resp, token)
	return nil
}

func (*LoginApp) Logout(req api.LogoutReq, success *bool) error {
	err := domain_service.TokenDomainService.Logout(req.Token)
	if err != nil {
		return err
	}
	*success = true
	return nil
}

func (*LoginApp) RefreshToken(req api.RefreshTokenReq, resp *api.TokenResp) error {
	token, err := domain_service.TokenDomainService.RefreshToken(req.Token, req.RefreshToken)
	if err != nil {
		return err
	}
	return copier.Copy(resp, token)
}
