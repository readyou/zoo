package domain

import (
	"time"
	"zoo/util"
)

type Token struct {
	Username          string
	Token             string
	ExpireTime        int64
	RefreshToken      string
	RefreshExpireTime int64
}

func NewToken(username string) *Token {
	// 默认token有效期1天，refreshToken有效期3天
	return NewTokenWithTime(username, time.Hour*24, time.Hour*24*3)
}

func NewTokenWithTime(username string, expireTime, refreshExpireTime time.Duration) *Token {
	return &Token{
		Username:          username,
		Token:             util.Encrypt.RandStr(),
		RefreshToken:      util.Encrypt.RandStr(),
		ExpireTime:        time.Now().Add(expireTime).Unix(),
		RefreshExpireTime: time.Now().Add(refreshExpireTime).Unix(),
	}
}

func (token *Token) IsRefreshTokenExpired() bool {
	return time.Now().Unix() > token.RefreshExpireTime
}

func (token *Token) IsTokenExpired() bool {
	return time.Now().Unix() > token.ExpireTime
}

func (token *Token) GetRefreshExpireDuration() time.Duration {
	t := time.Unix(token.RefreshExpireTime, 0)
	return t.Sub(time.Now())
}

func GetTokenKey(token string) string {
	return "LOGIN_TOKEN_" + token
}
