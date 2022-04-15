package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain/domain_service"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"github.com/jinzhu/copier"
)

type UserApp struct {
}

func (u *UserApp) Register(req api.RegisterReq, resp *api.ProfileResp) error {
	user := &domain.User{}
	user.Username = req.Username
	if err := user.CheckUsername(); err != nil {
		return err
	}
	if err := user.CheckPassword(req.Password); err != nil {
		return err
	}
	user.HashedPassword = user.EncryptPassword(req.Password)
	if err := repository.UserRepository.Create(user); err != nil {
		return err
	}
	if err := u.getProfile(req.Username, resp); err != nil {
		return err
	}
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