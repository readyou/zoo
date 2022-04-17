package main

import (
	"log"
	tcp_starter "git.garena.com/xinlong.wu/zoo/tcp-server"
)

func main() {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	tcp_starter.StartTcpServer()
}
