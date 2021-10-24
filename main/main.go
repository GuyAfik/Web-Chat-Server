package main

import (
	"web-chat-server/chat"
	"flag"
)

var (
	port = flag.String("p", ":8080", "set port")
)

func init() {
	flag.Parse()
}

func main() {
	chat.Start(*port)
}
