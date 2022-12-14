package main

import (
	http_starter "git.garena.com/xinlong.wu/zoo/http-server"
	tcp_starter "git.garena.com/xinlong.wu/zoo/tcp-server"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	go tcp_starter.StartTcpServer()
	time.Sleep(time.Second)
	http_starter.StartHttpServer()
}
