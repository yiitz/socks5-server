# socks5-server

A socks5 over tls server(tcp/udp) written in golang.

[![Travis](https://travis-ci.com/net-byte/socks5-server.svg?branch=main)](https://github.com/net-byte/socks5-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/net-byte/socks5-server)](https://goreportcard.com/report/github.com/net-byte/socks5-server)
![image](https://img.shields.io/badge/License-MIT-orange)
![image](https://img.shields.io/badge/License-Anti--996-red)

# Usage
```
Usage of /main:
  -l string
        local address (default "127.0.0.1:1080")
  -p string
        password
  -u string
        username
  -sk string
        server key file path (default "../certs/server.key")
  -sp string
        server pem file path (default "../certs/server.pem")
  -tls
        enable tls
```

# Docker

## Run server
```
docker run  -d --restart=always \
-p 1080:1080 -p 1080:1080/udp --name socks5-server netbyte/socks5-server -l :1080
```

## Run server with auth
```
docker run  -d --restart=always \
-p 1080:1080 -p 1080:1080/udp --name socks5-server netbyte/socks5-server -l :1080 -u root -p 123456
```

## Run server over tls with auth
```
docker run  -d --restart=always \
-p 1080:1080 -p 1080:1080/udp --name socks5-server netbyte/socks5-server -l :1080 -u root -p 123456 -tls -sk /app/certs/server.key -sp /app/certs/server.pem
```

# License
[The MIT License (MIT)](https://raw.githubusercontent.com/net-byte/opensocks/main/LICENSE)
