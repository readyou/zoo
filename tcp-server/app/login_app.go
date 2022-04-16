package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain/domain-service"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"github.com/jinzhu/copier"
	"time"
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

func (*LoginApp) EchoTokenForTest(req api.EchoTokenReq, resp *api.TokenResp) error {
	if req.Sleep > 0 {
		time.Sleep(time.Duration(req.Sleep) * time.Millisecond)
	}
	token := domain.NewToken(req.Token)
	if req.UseDB {
		m, err := infra.XDB.QueryString("select now() as n")
		//log.Printf("%#v, %s\n", m, err)
		if err != nil {
			return err
		}
		token.RefreshToken = m[0]["n"]
	}
	resp.Token = req.Token
	resp.ExpireTime = token.ExpireTime
	resp.RefreshToken = token.RefreshToken
	resp.RefreshExpireTime = token.RefreshExpireTime
	return nil
}
