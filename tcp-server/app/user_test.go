package app

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"os"
	"strings"
	"sync"
	"testing"
	"git.garena.com/xinlong.wu/zoo/api"
	infra2 "git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/util"
)

func TestMain(m *testing.M) {
	infra2.InitXDB()
	infra2.InitDB()
	infra2.InitRedis()
	os.Exit(m.Run())
}

var userApp = &UserApp{}

func TestUserApp_Register(t *testing.T) {
	req := api.RegisterReq{
		Password: util.UUID.NewString(),
	}
	resp := new(bool)
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
	assert.True(t, *resp)

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
	assert.True(t, *resp)
}

func TestUserApp_UpdateProfile(t *testing.T) {
	resp := RegisterAndLogin(t)
	profile := &api.ProfileResp{}
	userApp.getProfile(resp.Username, profile)
	assert.Equal(t, "", profile.Avatar)
	assert.Equal(t, "", profile.Nickname)

	// 修改avatar和nickname
	avatar := "https://www.shopee.com/" + resp.Token
	nickname := "nickname"
	userApp.UpdateProfile(api.UpdateProfileReq{
		Token:    resp.Token,
		Nickname: nickname,
		Avatar:   avatar,
	}, profile)
	assert.Equal(t, avatar, profile.Avatar)
	assert.Equal(t, nickname, profile.Nickname)

	profile = &api.ProfileResp{}
	userApp.getProfile(resp.Username, profile)
	assert.Equal(t, avatar, profile.Avatar)
	assert.Equal(t, nickname, profile.Nickname)
}

func Register(username, password string) (*bool, error) {
	req := api.RegisterReq{
		Username: username,
		Password: password,
	}
	resp := new(bool)
	if err := userApp.Register(req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func TestUserApp_Register_bench(t *testing.T) {
	util.GetQPS(func(n int) {
		wg := sync.WaitGroup{}
		wg.Add(n)
		sem := semaphore.NewWeighted(100)
		for i := 0; i < n; i++ {
			go func() {
				sem.Acquire(context.Background(), 1)
				defer func() {
					wg.Done()
					sem.Release(1)
				}()
				username := util.UUID.NewString()
				password := util.UUID.NewString()
				_, err := Register(username, password)
				assert.Nil(t, err, err)
			}()
		}
		wg.Wait()
	}, 100000)
}
