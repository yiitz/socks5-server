#!bin/bash
export GO111MODULE=on

#Linux amd64
GOOS=linux GOARCH=amd64 go build -o ./bin/socks5-server-linux-amd64 ./main.go
#Linux arm64
GOOS=linux GOARCH=arm64 go build -o ./bin/socks5-server-linux-arm64 ./main.go
#Mac amd64
GOOS=darwin GOARCH=amd64 go build -o ./bin/socks5-server-darwin-amd64 ./main.go
#Windows amd64
GOOS=windows GOARCH=amd64 go build -o ./bin/socks5-server-windows-amd64.exe ./main.go
#Operwrt amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o ./bin/socks5-server-openwrt-amd64 ./main.go

echo "DONE!!!"
