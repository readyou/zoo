package repository

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
)

func TestMain(m *testing.M) {
	infra.InitRedis()
	os.Exit(m.Run())
}

func TestTokenRepository_Save_expireTime(t *testing.T) {
	username := util.UUID.NewString()
	token := domain.NewTokenWithTime(username, time.Second*1, time.Second*4)
	err := TokenRepository.Save(token)
	assert.Nil(t, err, err)

	time.Sleep(time.Second * 2)
	token2, err := TokenRepository.Get(token.Token)
	assert.Nil(t, err, err)
	assert.Equal(t, token.Token, token2.Token)
	assert.True(t, token2.IsTokenExpired())
	assert.False(t, token2.IsRefreshTokenExpired())

	time.Sleep(time.Second * 2)
	token3, err := TokenRepository.Get(token.Token)
	assert.NotNil(t, err)
	assert.Nil(t, token3)
	assert.Contains(t, err.Error(), err_const.LoginRequired)
}
