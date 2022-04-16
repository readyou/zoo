package main

import (
	"git.garena.com/xinlong.wu/zoo/http-server/starter"
	starter2 "git.garena.com/xinlong.wu/zoo/tcp-server/starter"
	"time"
)

func main() {
	go starter2.StartTcpServer()
	time.Sleep(time.Second)
	starter.StartHttpServer()
}
