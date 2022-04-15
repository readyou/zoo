package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/config"
	"git.garena.com/xinlong.wu/zoo/http-server/starter"
	starter2 "git.garena.com/xinlong.wu/zoo/tcp-server/starter"
	"git.garena.com/xinlong.wu/zoo/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var client = &http.Client{}

func TestMain(m *testing.M) {
	go starter2.StartTcpServer()
	time.Sleep(time.Second)
	go starter.StartHttpServer()
	time.Sleep(time.Second)
	os.Exit(m.Run())
}

func username(i int) string {
	return fmt.Sprintf("user_%d", i)
}

func TestUserRegister(t *testing.T) {
	n := 1000
	//n := 2
	GetQPS(func() {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		sem := semaphore.NewWeighted(200)
		for i := 0; i < n; i++ {
			go func(seq int) {
				sem.Acquire(context.Background(), 1)
				register(t, seq)
				sem.Release(1)
				wg.Done()
			}(i)
		}
		wg.Wait()
	}, n)
}

func GetQPS(f func(), n int) {
	start := time.Now()
	f()
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	log.Printf("Qps: %f\n", float64(n)/seconds)
}

func getURL(path string) string {
	return fmt.Sprintf("http://%s%s", config.Config.HTTPServer.Address, path)
}

func register(t *testing.T, seq int) {
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
	bytes, err := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(bytes))

	assert.Nil(t, err)

	result := &util.RequestResult{}
	err = json.Unmarshal(bytes, result)
	assert.Nil(t, err, err)
	assert.True(t, result.Success)
	assert.Equal(t, username, (result.Data.(map[string]any)["Username"]))

}
