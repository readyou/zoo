package app

import (
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/infra"
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	infra.InitDB()
	infra.InitRedis()
	os.Exit(m.Run())
}

var userApp = &UserApp{}

func TestUserApp_Register(t *testing.T) {
	req := api.RegisterReq{
		Password: util.UUID.NewString(),
	}
	resp := &api.ProfileResp{}
	// 用户名为空
	err := userApp.Register(req, resp)
	assertErrContains(t, err, "username", "length")

	// 用户名超长
	req.Username = strings.Repeat("1", 65)
	err = userApp.Register(req, resp)
	assertErrContains(t, err, "username", "length")

	// 用户名字符不合法
	req.Username = "..."
	err = userApp.Register(req, resp)
	assertErrContains(t, err, "username", "letters")

	// 密码为空
	req.Username = util.UUID.NewString()
	req.Password = ""
	err = userApp.Register(req, resp)
	assertErrContains(t, err, "password", "length")

	// 密码超长
	req.Password = strings.Repeat("1", 65)
	err = userApp.Register(req, resp)
	assertErrContains(t, err, "password", "length")

	// 正常注册成功
	req = api.RegisterReq{
		Username: util.UUID.NewString(),
		Password: util.UUID.NewString(),
	}
	err = userApp.Register(req, resp)
	assert.Nil(t, err, err)
	assert.Equal(t, req.Username, resp.Username)
	assert.True(t, resp.Id > 0)

	// 再次用相同的username注册，期待报错
	err = userApp.Register(req, resp)
	assert.NotNil(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "duplicate")
}

func assertErrContains(t *testing.T, err error, strs ...string) {
	assert.NotNil(t, err)
	for _, str := range strs {
		assert.Contains(t, err.Error(), str)
	}
}

func TestUserApp_GetProfile(t *testing.T) {
	username := util.UUID.NewString()
	password := util.UUID.NewString()
	resp, err := Register(username, password)
	assert.Nil(t, err, err)
	assert.Equal(t, username, resp.Username)
	assert.True(t, resp.Id > 0)
}

func TestUserApp_UploadImg(t *testing.T) {

}

func TestUserApp_UpdateProfile(t *testing.T) {
	resp := RegisterAndLogin(t)

	profile := &api.ProfileResp{}
	avatar := "https://www.shopee.com/" + resp.Token
	userApp.UpdateProfile(api.UpdateProfileReq{
		Token:    resp.Token,
		Nickname: resp.Token,
		Avatar:   avatar,
	}, profile)
	assert.Equal(t, avatar, profile.Avatar)
	assert.Equal(t, resp.Token, profile.Nickname)

	profile = &api.ProfileResp{}
	userApp.getProfile(resp.Username, profile)
	assert.Equal(t, avatar, profile.Avatar)
	assert.Equal(t, resp.Token, profile.Nickname)
}

func Register(username, password string) (*api.ProfileResp, error) {
	req := api.RegisterReq{
		Username: username,
		Password: password,
	}
	resp := &api.ProfileResp{}
	if err := userApp.Register(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
