package http_server

import (
	"fmt"
	"github.com/jinzhu/copier"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
	"git.garena.com/xinlong.wu/zoo/ant"
	"git.garena.com/xinlong.wu/zoo/api"
	"git.garena.com/xinlong.wu/zoo/bee"
	"git.garena.com/xinlong.wu/zoo/config"
	"git.garena.com/xinlong.wu/zoo/tcp-server/domain"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra/err_const"
	"git.garena.com/xinlong.wu/zoo/util"
)

const (
	UploadBaseDir = "/opt/data/upload"
	ImageDirName  = "img"
)

var imageTypes = []string{"bmp", "jpg", "png", "tif", "gif", "pcx", "tga", "exif", "fpx", "svg", "psd", "cdr", "pcd", "dxf", "ufo", "eps", "ai", "raw", "WMF", "webp", "avif", "apng"}

type proxyItem struct {
	HTTPPath      string // http.URL.Path
	ServiceMethod string // rpc ServiceMethod
	ArgType       reflect.Type
	ReplyType     reflect.Type
}

func StartHttpServer() {
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
		{"/api/user/register", "UserApp.Register", reflect.TypeOf(api.RegisterReq{}), reflect.TypeOf(new(bool))},
		{"/api/user/update", "UserApp.UpdateProfile", reflect.TypeOf(api.UpdateProfileReq{}), reflect.TypeOf(&api.ProfileResp{})},
		{"/api/user/get", "UserApp.GetProfile", reflect.TypeOf(api.GetProfileReq{}), reflect.TypeOf(&api.ProfileResp{})},
		{"/api/user/login", "LoginApp.Login", reflect.TypeOf(api.LoginReq{}), reflect.TypeOf(&api.TokenResp{})},
		{"/api/user/logout", "LoginApp.Logout", reflect.TypeOf(api.LogoutReq{}), reflect.TypeOf(new(bool))},
		{"/api/user/token/validate", "LoginApp.ValidateToken", reflect.TypeOf(api.ValidateTokenReq{}), reflect.TypeOf(&api.TokenResp{})},
		{"/api/user/token/refresh", "LoginApp.RefreshToken", reflect.TypeOf(api.RefreshTokenReq{}), reflect.TypeOf(&api.TokenResp{})},
		{"/api/user/token/echo", "LoginApp.EchoTokenForTest", reflect.TypeOf(api.EchoTokenReq{}), reflect.TypeOf(&api.TokenResp{})},
	}
	for _, item := range proxyList {
		i := item
		// 约定：统一使用POST请求, http-server只是转发请求
		engine.Post(item.HTTPPath, func(c *ant.Context) {
			proxyToRPCServer(rpcClient, c, i)
		})
	}

	engine.Get("/api/ping", func(context *ant.Context) {
		context.JSON(http.StatusOK, util.Result.OK("OK"))
	})
	engine.Post("/api/image/upload", func(context *ant.Context) {
		_, err := validateToken(rpcClient, context)
		if err != nil {
			return
		}
		handlerFileUpload(context, imageTypes, 2000_000)
	})

	engine.Run(config.Config.HTTPServer.Address)
}

func handlerFileUpload(context *ant.Context, fileTypes []string, maxBytes int64) {
	request := context.Request
	request.ParseMultipartForm(maxBytes) // limit your max input length!
	file, header, err := request.FormFile("file")
	if err != nil {
		log.Printf("get FormFile error: %s\n", err)
		context.JSON(http.StatusOK, util.Result.Fail(err_const.InvalidParam, "file not found"))
		return
	}
	defer file.Close()
	fmt.Printf("upload file, filename=%s\n", header.Filename)
	nameParts := strings.Split(header.Filename, ".")
	if len(nameParts) < 2 || !containsIgnoreCase(fileTypes, nameParts[1]) {
		context.JSON(http.StatusOK, util.Result.Fail(err_const.InvalidParam, fmt.Sprintf("support image type: %s", strings.Join(fileTypes, ","))))
		return
	}

	filename := util.UUID.NewString() + "." + nameParts[1]
	f, err := os.OpenFile(UploadBaseDir+"/"+ImageDirName+"/"+filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		context.JSON(http.StatusOK, util.Result.Fail(err_const.SystemError, err.Error()))
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		panic("create image file error")
	}
	context.JSON(http.StatusOK, util.Result.OK(getImageUrl(filename)))
}

func getImageUrl(filename string) string {
	return ImageDirName + "/" + filename
}

func containsIgnoreCase(strList []string, str string) bool {
	str = strings.ToLower(str)
	for _, v := range strList {
		if strings.ToLower(v) == str {
			return true
		}
	}
	return false
}

// 校验登录token，如果校验失败，直接将结果写入context
func validateToken(rpc *bee.Client, context *ant.Context) (token *domain.Token, err error) {
	tokenStr := context.Request.Header.Get("Token")
	args := &api.ValidateTokenReq{
		Token: tokenStr,
	}

	reply := &api.TokenResp{}
	if err = rpc.Call("UserApp.ValidateToken", args, reply); err != nil {
		log.Printf("rpc request error: %s\n", err.Error())
		if strings.HasPrefix(err.Error(), err_const.LoginRequired) {
			context.JSON(http.StatusOK, util.Result.Fail(err_const.LoginRequired, ""))
			return
		}
		log.Printf("rpc validate token error: %s\n", err.Error())
		context.JSON(http.StatusOK, util.Result.Fail(err_const.SystemError, err.Error()))
		return
	}

	token = &domain.Token{}
	copier.Copy(token, reply)
	return
}

func proxyToRPCServer(client *bee.Client, c *ant.Context, item proxyItem) {
	args := reflect.New(item.ArgType).Interface()
	reply := reflect.New(item.ReplyType).Interface()
	c.ParseJSON(&args)
	if err := client.Call(item.ServiceMethod, args, reply); err != nil {
		log.Printf("rpc request error: %s\n", err.Error())
		c.JSON(http.StatusOK, util.Result.Fail("SystemError", err.Error()))
		return
	}

	c.JSON(http.StatusOK, util.Result.OK(reply))
}
