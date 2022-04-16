package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain/domain-service"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"github.com/jinzhu/copier"
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

func (u *UserApp) UpdateProfile(req api.UpdateProfileReq, resp *api.ProfileResp) error {
	token, err := domain_service.TokenDomainService.ValidateToken(req.Token)
	if err != nil {
		return err
	}

	user, err := repository.UserRepository.GetUserByName(token.Username)
	if err != nil {
		return err
	}

	copier.Copy(&user, req)
	if err := repository.UserRepository.Update(user); err != nil {
		return err
	}

	if err := u.getProfile(user.Username, resp); err != nil {
		return err
	}
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
	user, err := repository.UserRepository.GetUserByName(username)
	if err != nil {
		return err
	}
	copier.Copy(resp, user)
	return nil
}

func (u *UserApp) UploadImg(req api.UploadImgReq, resp *string) error {

	return nil
}
