package tcp_server

import (
	"log"
	"git.garena.com/xinlong.wu/zoo/bee"
	"git.garena.com/xinlong.wu/zoo/config"
	"git.garena.com/xinlong.wu/zoo/tcp-server/app"
	"git.garena.com/xinlong.wu/zoo/tcp-server/infra"
)

func StartTcpServer() {
	// 初始化
	infra.InitXDB()
	//infra.InitDB()
	infra.InitRedis()

	// 注册RPC
	server := bee.NewServer()
	registerService(server, new(app.UserApp))
	registerService(server, new(app.LoginApp))

	// 启动服务
	address := config.Config.TcpServer.Address
	//server.StartHTTP(address)
	server.Start(address)
	log.Println("http server started")
}

func registerService(server *bee.Server, service any) {
	err := server.Register(service)
	if err != nil {
		log.Fatalf("registor error, %s: %s", service, err)
	}
}
