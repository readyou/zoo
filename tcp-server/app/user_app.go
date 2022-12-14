package app

import (
	"github.com/jinzhu/copier"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain/domain-service"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
)

type UserApp struct {
}

func (u *UserApp) Register(req api.RegisterReq, resp *bool) error {
	if err := domain_service.UserDomainService.Register(req.Username, req.Password); err != nil {
		return err
	}
	*resp = true
	return nil
}

func (u *UserApp) ValidateToken(req api.ValidateTokenReq, resp *api.TokenResp) error {
	token, err := domain_service.TokenDomainService.ValidateToken(req.Token)
	if err != nil {
		return err
	}
	copier.Copy(resp, token)
	return nil
}

func (u *UserApp) UpdateProfile(req api.UpdateProfileReq, resp *api.ProfileResp) error {
	token, err := domain_service.TokenDomainService.ValidateToken(req.Token)
	if err != nil {
		return err
	}
	user, err := domain_service.UserDomainService.UpdateProfile(token.Username, req.Nickname, req.Avatar)
	if err != nil {
		return err
	}
	copier.Copy(resp, user)
	return nil
}

func (u *UserApp) GetProfile(req api.GetProfileReq, resp *api.ProfileResp) error {
	token, err := domain_service.TokenDomainService.ValidateToken(req.Token)
	if err != nil {
		return err
	}

	if err := u.getProfile(token.Username, resp); err != nil {
		return err
	}
	return nil
}

func (u *UserApp) getProfile(username string, resp *api.ProfileResp) error {
	user, err := repository.UserRepository.GetCache(username)
	if err != nil {
		return err
	}
	copier.Copy(resp, user)
	return nil
}
