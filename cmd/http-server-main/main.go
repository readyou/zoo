package main

import (
	http_starter "git.garena.com/xinlong.wu/zoo/http-server"
	"log"
)

func main() {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	http_starter.StartHttpServer()
}
