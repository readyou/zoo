package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/bee"
	"git.garena.com/xinlong.wu/zoo/config"
	"git.garena.com/xinlong.wu/zoo/http-server/starter"
	starter2 "git.garena.com/xinlong.wu/zoo/tcp-server/starter"
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

var client = &http.Client{}

func startServer(m *testing.M) {
	go starter2.StartTcpServer()
	time.Sleep(time.Second)
	go starter.StartHttpServer()
	time.Sleep(time.Second)
	//os.Exit(m.Run())
}

func TestUserRegister(t *testing.T) {
	startServer(&testing.M{})
	n := 10000
	//n := 2
	util.GetQPS(func(n int) {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		sem := semaphore.NewWeighted(150)
		for i := 0; i < n; i++ {
			go func(seq int) {
				sem.Acquire(context.Background(), 1)
				defer func() {
					wg.Done()
					sem.Release(1)
				}()
				//register(t)
				//echoToken(t)
			}(i)
		}
		wg.Wait()
	}, n)
}

func getURL(path string) string {
	return fmt.Sprintf("http://%s%s", config.Config.HTTPServer.Address, path)
}

func TestPing(t *testing.T) {
	startServer(&testing.M{})
	util.GetQPSAsnyc(func(i int) {
		//resp, err := http.Get(getURL("/api/ping"))
		//assert.Nil(t, err, err)
		//assert.NotNil(t, resp)
		//assertResponseSuccess(t, resp)
	}, 1000, 100)
}

func assertResponseSuccess(t *testing.T, resp *http.Response) {
	bytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(bytes), "\"Success\":true"))
}

func echoToken(t *testing.T) {
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

	//result := &util.RequestResult{}
	//err = json.Unmarshal(bytes, result)
	//assert.Nil(t, err, err)
	//assert.True(t, result.Success)
	//assert.Equal(t, token, (result.Data.(map[string]any)["Token"]))

}

func register(t *testing.T) {
	username := util.UUID.NewString()
	args := api.RegisterReq{
		Username: username,
		Password: username,
	}
	jsonStr, err := json.Marshal(args)
	req, err := http.NewRequest("POST", getURL("/api/user/register"), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Nil(t, err)
}

func TestRPCEchoToken(t *testing.T) {
	go starter2.StartTcpServer()
	clientsCount := 10
	clients := make([]*bee.Client, clientsCount)
	for i := 0; i < clientsCount; i++ {
		clients[i] = bee.NewClient()
		err := clients[i].Dial(config.Config.TcpServer.Address)
		assert.Nil(t, err, err)
	}
	time.Sleep(time.Second * 2)

	n := 100000
	//n := 2
	util.GetQPS(func(n int) {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		sem := semaphore.NewWeighted(300)
		for j := 0; j < n/clientsCount; j++ {
			for i := 0; i < clientsCount; i++ {
				go func(i, j int) {
					sem.Acquire(context.Background(), 1)
					defer func() {
						sem.Release(1)
						wg.Done()
					}()
					testEchoToken(t, clients[i])
				}(i, j)
			}
		}
		wg.Wait()
	}, n)

}

func testAdd(t *testing.T, client *bee.Client) {

}

func testEchoToken(t *testing.T, client *bee.Client) {
	token := util.UUID.NewString()
	args := api.RefreshTokenReq{
		Token:        token,
		RefreshToken: token,
	}
	reply := &api.TokenResp{}
	err := client.Call("LoginApp.EchoTokenForTest", args, reply)
	assert.Nil(t, err, err)
	assert.Equal(t, token, reply.Token)
}

func testRegister(t *testing.T, client *bee.Client) {
	username := util.UUID.NewString()
	args := api.RegisterReq{
		Username: username,
		Password: username,
	}
	reply := &api.ProfileResp{}
	err := client.Call("UserApp.Register", args, reply)
	assert.Nil(t, err, err)
	assert.Equal(t, username, reply.Username)
}
