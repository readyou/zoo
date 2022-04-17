package app

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/tcp-server/repository"
	"git.garena.com/xinlong.wu/zoo/util"
)

var loginApp = &LoginApp{}

func TestLoginApp_Login_failed(t *testing.T) {
	username := util.UUID.NewString()
	password := util.UUID.NewString()
	_, err := Register(username, password)
	assert.Nil(t, err, err)

	req := api.LoginReq{
		Username: util.UUID.NewString(),
		Password: password,
	}
	token := &api.TokenResp{}
	err = loginApp.Login(req, token)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err_const.UserNotExists)

	req.Username = username
	req.Password = "invalidpassword"
	token = &api.TokenResp{}
	err = loginApp.Login(req, token)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err_const.InvalidPassword)
}

func TestLoginApp_Login_success(t *testing.T) {
	RegisterAndLogin(t)
}

func RegisterAndLogin(t *testing.T) *api.TokenResp {
	username := util.UUID.NewString()
	password := util.UUID.NewString()
	success, err := Register(username, password)
	assert.Nil(t, err, err)
	assert.True(t, *success)

	req := api.LoginReq{
		Username: username,
		Password: password,
	}
	token := &api.TokenResp{}
	err = loginApp.Login(req, token)
	assert.Nil(t, err, err)
	assert.Equal(t, username, token.Username)

	token2, err := repository.TokenRepository.Get(token.Token)
	assert.Nil(t, err, err)
	assert.Equal(t, token.Token, token2.Token)
	return token
}

func TestLoginApp_Logout(t *testing.T) {
	resp := RegisterAndLogin(t)
	token := resp.Token
	success := false
	err := loginApp.Logout(api.LogoutReq{Token: token}, &success)
	assert.Nil(t, err, err)
	assert.True(t, success)

	_, err = repository.TokenRepository.Get(token)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "LoginRequired")
}

func TestLoginApp_RefreshToken_success(t *testing.T) {
	username := util.UUID.NewString()
	token := domain.NewTokenWithTime(username, time.Second*1, time.Second*3)
	err := repository.TokenRepository.Save(token)
	assert.Nil(t, err, err)

	time.Sleep(time.Second * 1)
	resp := &api.TokenResp{}
	err = loginApp.RefreshToken(api.RefreshTokenReq{
		Token:        token.Token,
		RefreshToken: token.RefreshToken,
	}, resp)
	assert.Nil(t, err, err)
	assert.NotEqual(t, resp.Token, token.Token)
	assert.NotEqual(t, resp.RefreshToken, token.RefreshToken)

	token2, err := repository.TokenRepository.Get(resp.Token)
	assert.Nil(t, err, err)
	assert.Equal(t, resp.Token, token2.Token)
}

func TestLoginApp_RefreshToken_fail_invalidRefreshToken(t *testing.T) {
	username := util.UUID.NewString()
	token := domain.NewTokenWithTime(username, time.Second*1, time.Second*5)

	// RefreshToken不对，刷新失败
	err := repository.TokenRepository.Save(token)
	assert.Nil(t, err, err)

	assertRefreshFail(t, api.RefreshTokenReq{
		Token:        token.Token,
		RefreshToken: "",
	})
}

func TestLoginApp_RefreshToken_fail_refreshTokenExpired(t *testing.T) {
	username := util.UUID.NewString()
	token := domain.NewTokenWithTime(username, time.Second*1, time.Second*3)
	err := repository.TokenRepository.Save(token)
	assert.Nil(t, err, err)

	// 过期，刷新失败
	time.Sleep(time.Second * 4)
	assertRefreshFail(t, api.RefreshTokenReq{
		Token:        token.Token,
		RefreshToken: token.RefreshToken,
	})
}

func assertRefreshFail(t *testing.T, req api.RefreshTokenReq) {
	resp := &api.TokenResp{}
	err := loginApp.RefreshToken(req, resp)
	assert.NotNil(t, err, err)
	assert.Contains(t, err.Error(), "LoginRequired")

	// refresh失败，会删除原来的token（为了安全），所以这里无法再获取到token
	_, err = repository.TokenRepository.Get(req.Token)
	assert.NotNil(t, err, err)
	assert.Contains(t, err.Error(), "LoginRequired")
}
