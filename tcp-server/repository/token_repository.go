package repository

import (
	"context"
	"encoding/json"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var TokenRepository *tokenRepository = &tokenRepository{}

type tokenRepository struct {
}

func (*tokenRepository) Save(token *domain.Token) (err error) {
	bytes, err := json.Marshal(token)
	if err != nil {
		return
	}
	js := string(bytes)
	statusCmd := infra.RDB.SetEX(context.Background(), domain.GetTokenKey(token.Token), js, time.Duration(token.RefreshExpireTime))
	if err = statusCmd.Err(); err != nil {
		log.Printf("save token to redis error:%s\n", err.Error())
	}
	return
}

func (*tokenRepository) Get(token string) (tok *domain.Token, err error) {
	key := domain.GetTokenKey(token)
	val, err := infra.RDB.Get(context.Background(), key).Result()
	switch {
	case err == redis.Nil:
		err = util.Err.ServerError(err_const.LoginRequired, "please login first")
	case err != nil:
		// get failed
		err = util.Err.ServerError(err_const.SystemError, "get token error")
	default:
		tok = &domain.Token{}
		err = json.Unmarshal([]byte(val), tok)
	}
	return
}

func (*tokenRepository) Delete(token string) error {
	key := domain.GetTokenKey(token)
	_, err := infra.RDB.Del(context.Background(), key).Result()
	if err != nil {
		log.Printf("delete token error: %s\n", err.Error())
		return util.Err.ServerError(err_const.SystemError, "get token error")
	}
	return nil
}
