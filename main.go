package main

import (
	"zoo/http-server/starter"
	starter2 "zoo/tcp-server/starter"
	"time"
)

func main() {
	go starter2.StartTcpServer()
	time.Sleep(time.Second)
	starter.StartHttpServer()
}
