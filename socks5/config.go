package socks5

type Config struct {
	LocalAddr string
	Username  string
	Password  string
	ServerKey string
	ServerPem string
	TLS       bool
}
