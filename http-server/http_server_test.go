package http_server

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/config"
	tcp_starter "git.garena.com/xinlong.wu/zoo/tcp-server"
	"git.garena.com/xinlong.wu/zoo/util"
)

import (
	"encoding/json"
)

// 此测试文件只测试基本的http请求是否正确，以及RPC转发请求是否响应正确;
// 不测试tcp-server的具体逻辑
var client = &http.Client{}

func TestMain(m *testing.M) {
	util.ConfigLog()
	go tcp_starter.StartTcpServer()
	time.Sleep(time.Second)
	go StartHttpServer()
	time.Sleep(time.Second)
	os.Exit(m.Run())
}

func getURL(path string) string {
	return fmt.Sprintf("http://%s%s", config.Config.HTTPServer.Address, path)
}

func TestPing(t *testing.T) {
	resp, err := http.Get(getURL("/api/ping"))
	assert.Nil(t, err, err)
	assert.NotNil(t, resp)
	assertResponseSuccess(t, resp)
}

func assertResponseSuccess(t *testing.T, resp *http.Response) {
	bytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(bytes), "\"Success\":true"))
}

func TestEchoToken(t *testing.T) {
	token := util.UUID.NewString()
	args := api.RefreshTokenReq{
		Token:        token,
		RefreshToken: token,
	}
	js, err := json.Marshal(args)
	req, err := http.NewRequest("POST", getURL("/api/user/token/echo"), bytes.NewBuffer(js))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Nil(t, err)
	bytes, err := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(bytes))

	assert.Nil(t, err)

	result := &util.RequestResult{}
	err = json.Unmarshal(bytes, result)
	assert.Nil(t, err, err)
	assert.True(t, result.Success)
	assert.Equal(t, token, (result.Data.(map[string]any)["Token"]))
}
