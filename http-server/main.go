package main

import (
	"git.garena.com/xinlong.wu/zoo/ant"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/bee"
	"git.garena.com/xinlong.wu/zoo/config"
	"git.garena.com/xinlong.wu/zoo/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
	"log"
	"net/http"
	"reflect"
	"time"
)

type proxyItem struct {
	HTTPPath      string // http.URL.Path
	ServiceMethod string // rpc ServiceMethod
	ArgType       reflect.Type
	ReplyType     reflect.Type
}

func main() {
	rpcClient := startRPCClient()
	startHttpServer(rpcClient)
}

// TODO: 实现连接池
func startRPCClient() *bee.Client {
	// 初始化RPC Client
	rpcClient := bee.NewClientWithOption(bee.ClientOption{
		ConnectTimeout:  time.Second * 10,
		ResponseTimeout: time.Second * 10,
	})
	if err := rpcClient.Dial(config.Config.TcpServer.Address); err != nil {
		log.Fatalln("start rpc client error", err)
	}
	return rpcClient
}

func startHttpServer(rpcClient *bee.Client) {
	engine := ant.Default()
	proxyList := []proxyItem{
		{"/api/user/register", "UserApp.Register", reflect.TypeOf(api.RegisterReq{}), reflect.TypeOf(&api.ProfileResp{})},
		{"/api/user/update", "UserApp.UpdateProfile", reflect.TypeOf(api.UpdateProfileReq{}), reflect.TypeOf(&api.ProfileResp{})},
		{"/api/user/get", "UserApp.GetProfile", reflect.TypeOf(api.GetProfileReq{}), reflect.TypeOf(&api.ProfileResp{})},
		{"/api/user/login", "LoginApp.Login", reflect.TypeOf(api.LoginReq{}), reflect.TypeOf(&api.TokenResp{})},
		{"/api/user/logout", "LoginApp.Logout", reflect.TypeOf(api.LogoutReq{}), reflect.TypeOf(new(bool))},
		{"/api/user/token/refresh", "LoginApp.RefreshToken", reflect.TypeOf(api.RefreshTokenReq{}), reflect.TypeOf(&api.TokenResp{})},
	}
	for _, item := range proxyList {
		i := item
		// 约定：统一使用POST请求, http-server只是转发请求
		engine.Post(item.HTTPPath, func(c *ant.Context) {
			proxy(rpcClient, c, i)
		})
	}

	// TODO: upload image

	engine.Run(config.Config.HTTPServer.Address)
}

func proxy(client *bee.Client, c *ant.Context, item proxyItem) {
	args := reflect.New(item.ArgType).Interface()
	reply := reflect.New(item.ReplyType).Interface()
	c.ParseJSON(&args)
	if err := client.Call(item.ServiceMethod, args, reply); err != nil {
		log.Printf("rpc request error: %s\n", err.Error())
		c.JSON(http.StatusOK, util.Result.Fail(err_const.SystemError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, util.Result.Succeed(reply))
}
