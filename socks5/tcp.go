package socks5

import (
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

func TCPProxy(conn net.Conn, data []byte) {
	host, port := getAddr(data)
	if host == "" || port == "" {
		return
	}
	remoteConn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Duration(Timeout)*time.Second)
	if err != nil {
		log.Printf("[tcp] failed to dial tcp %v", err)
		responseTCP(conn, ConnectionRefused)
		return
	}
	responseTCP(conn, SuccessReply)
	go copy(remoteConn, conn)
	go copy(conn, remoteConn)
}

func copy(to io.WriteCloser, from io.ReadCloser) {
	defer to.Close()
	defer from.Close()
	io.Copy(to, from)
}

func getAddr(b []byte) (host string, port string) {
	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	len := len(b)
	switch b[3] {
	case Ipv4Address:
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case FqdnAddress:
		host = string(b[5 : len-2])
	case Ipv6Address:
		host = net.IP(b[4:20]).String()
	default:
		return "", ""
	}
	port = strconv.Itoa(int(b[len-2])<<8 | int(b[len-1]))
	return host, port
}
