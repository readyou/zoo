package infra

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"git.garena.com/xinlong.wu/zoo/util"
)

type Adder struct {
	A int
	B int
}

func loadValue(key string, ret any) error {
	value := ret.(*Adder)
	value.A = 10
	value.B = 20
	return nil
}

func TestMain(m *testing.M) {
	InitRedis()
	os.Exit(m.Run())
}

func TestRds_GetOrSet(t *testing.T) {
	key := util.UUID.NewString()
	value := &Adder{}
	err := RedisUtil.GetOrSet(key, loadValue, time.Second, value)
	assert.Nil(t, err, err)
	assert.Equal(t, 10, value.A)
	assert.Equal(t, 20, value.B)

	// set and then get
	err = RedisUtil.SetEX(key, &Adder{
		A: 3,
		B: 4,
	}, time.Second)
	assert.Nil(t, err, err)

	err = RedisUtil.GetOrSet(key, loadValue, time.Second, value)
	assert.Nil(t, err, err)
	assert.Equal(t, 3, value.A)
	assert.Equal(t, 4, value.B)
}
