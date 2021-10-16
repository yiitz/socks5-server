FROM golang:alpine

WORKDIR /app
COPY . /app
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN go build -o ./bin/socks5-server ./main.go

ENTRYPOINT ["./bin/socks5-server"]

