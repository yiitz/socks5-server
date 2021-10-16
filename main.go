package main

import (
	"flag"

	"github.com/net-byte/socks5-server/socks5"
)

func main() {
	config := socks5.Config{}
	flag.StringVar(&config.LocalAddr, "l", "127.0.0.1:1080", "local address")
	flag.StringVar(&config.Username, "u", "", "username")
	flag.StringVar(&config.Password, "p", "", "password")
	flag.Parse()

	socks5.StartServer(config)
}
